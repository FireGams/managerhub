package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/managerhub/managerhub/controller/internal/auth"
)

func (s *Server) handleListNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.Store.ListNodes(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db error")
		return
	}
	type view struct {
		ID       string   `json:"id"`
		Name     string   `json:"name"`
		Hostname string   `json:"hostname"`
		OS       string   `json:"os"`
		Arch     string   `json:"arch"`
		IP       string   `json:"ip"`
		AgentVer string   `json:"agent_version"`
		Status   string   `json:"status"`
		Online   bool     `json:"online"`
		LastSeen *string  `json:"last_seen_at"`
		Tags     []string `json:"tags"`
	}
	out := make([]view, 0, len(nodes))
	for _, n := range nodes {
		var ls *string
		if n.LastSeenAt != nil {
			s := n.LastSeenAt.UTC().Format(rFC3339)
			ls = &s
		}
		out = append(out, view{
			ID: n.ID, Name: n.Name, Hostname: n.Hostname, OS: n.OS, Arch: n.Arch,
			IP: n.IP, AgentVer: n.AgentVersion, Status: n.Status,
			Online: s.Hub.Online(n.ID), LastSeen: ls, Tags: n.Tags,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

const rFC3339 = "2006-01-02T15:04:05Z07:00"

func (s *Server) handleGetNode(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	n, err := s.Store.GetNode(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "node not found")
		return
	}
	m, _ := s.Store.LatestMetrics(r.Context(), id)
	writeJSON(w, http.StatusOK, map[string]any{"node": n, "metrics": m, "online": s.Hub.Online(id)})
}

func (s *Server) handleNodeMetrics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	m, err := s.Store.LatestMetrics(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	writeJSON(w, http.StatusOK, []any{m})
}

func (s *Server) handleDeleteNode(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := s.Store.Pool.Exec(r.Context(), `DELETE FROM nodes WHERE id=$1`, id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "delete failed")
		return
	}
	c, _ := auth.ClaimsFrom(r.Context())
	_ = s.Store.AppendAudit(r.Context(), c.Subject, "node.delete", id, "")
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
