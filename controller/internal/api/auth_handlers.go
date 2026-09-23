package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/managerhub/managerhub/controller/internal/auth"
	"github.com/managerhub/managerhub/shared/protocol"
)

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	c, _ := auth.ClaimsFrom(r.Context())
	writeJSON(w, http.StatusOK, map[string]string{"username": c.Subject, "role": c.Role})
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	u, err := s.Store.GetUserByUsername(r.Context(), req.Username)
	if err != nil || !auth.CheckPassword(u.PasswordHash, req.Password) {
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	tok, err := auth.IssueToken(s.Cfg.JWTSecret, u.Username, u.Role, 12*time.Hour)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "token issue failed")
		return
	}
	_ = s.Store.AppendAudit(r.Context(), u.Username, "auth.login", "", "")
	writeJSON(w, http.StatusOK, map[string]string{"token": tok, "role": u.Role, "username": u.Username})
}

type bootstrapReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// handleBootstrap creates the very first admin account when none exists.
func (s *Server) handleBootstrap(w http.ResponseWriter, r *http.Request) {
	n, err := s.Store.CountUsers(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db error")
		return
	}
	if n > 0 {
		writeErr(w, http.StatusConflict, "already bootstrapped")
		return
	}
	var req bootstrapReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || len(req.Password) < 8 {
		writeErr(w, http.StatusBadRequest, "username required and password >= 8 chars")
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "hash failed")
		return
	}
	u, err := s.Store.CreateUser(r.Context(), req.Username, hash, protocol.RoleAdmin)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "create user failed")
		return
	}
	_ = s.Store.AppendAudit(r.Context(), u.Username, "auth.bootstrap", u.Username, "")
	writeJSON(w, http.StatusCreated, map[string]string{"id": u.ID, "username": u.Username, "role": u.Role})
}

type enrollTokenReq struct {
	Label string `json:"label"`
	TTL   string `json:"ttl"`
}

func (s *Server) handleCreateEnrollToken(w http.ResponseWriter, r *http.Request) {
	c, _ := auth.ClaimsFrom(r.Context())
	var req enrollTokenReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	ttl := s.Cfg.EnrollTokenTTL
	if req.TTL != "" {
		if d, err := time.ParseDuration(req.TTL); err == nil {
			ttl = d
		}
	}
	raw, hash := newEnrollSecret()
	t, err := s.Store.CreateEnrollToken(r.Context(), hash, req.Label, c.Subject, ttl)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "create token failed")
		return
	}
	_ = s.Store.AppendAudit(r.Context(), c.Subject, "enroll_token.create", t.ID, req.Label)
	writeJSON(w, http.StatusCreated, map[string]any{
		"id": t.ID, "token": raw, "label": t.Label, "expires_at": t.ExpiresAt,
	})
}

func (s *Server) handleListEnrollTokens(w http.ResponseWriter, r *http.Request) {
	list, err := s.Store.ListEnrollTokens(r.Context(), 50)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleDeleteEnrollToken(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := s.Store.Pool.Exec(r.Context(), `DELETE FROM enroll_tokens WHERE id=$1`, id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "delete failed")
		return
	}
	c, _ := auth.ClaimsFrom(r.Context())
	_ = s.Store.AppendAudit(r.Context(), c.Subject, "enroll_token.delete", id, "")
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
