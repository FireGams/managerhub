package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/managerhub/managerhub/controller/internal/db/store"
	"github.com/managerhub/managerhub/shared/protocol"
)

type enrollReq struct {
	EnrollToken  string            `json:"enroll_token"`
	Name         string            `json:"name"`
	Hostname     string            `json:"hostname"`
	OS           string            `json:"os"`
	Arch         string            `json:"arch"`
	IP           string            `json:"ip"`
	AgentVersion string            `json:"agent_version"`
	Tags         []string          `json:"tags"`
	Capabilities map[string]string `json:"capabilities"`
}

type enrollResp struct {
	NodeID    string `json:"node_id"`
	NodeToken string `json:"node_token"`
}

// handleEnroll exchanges a one-time enrollment token for a node identity.
func (s *Server) handleEnroll(w http.ResponseWriter, r *http.Request) {
	var req enrollReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeErr(w, http.StatusBadRequest, "name is required")
		return
	}
	if _, err := s.Store.ConsumeEnrollToken(r.Context(), hashSecret(req.EnrollToken)); err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid or used enroll token")
		return
	}
	nodeToken, tokenHash := newEnrollSecret()
	nodeID := uuid.NewString()
	n, err := s.Store.UpsertNode(r.Context(), store.Node{
		ID: nodeID, Name: req.Name, Hostname: req.Hostname, OS: req.OS, Arch: req.Arch,
		IP: req.IP, AgentVersion: req.AgentVersion, Tags: req.Tags, Capabilities: req.Capabilities,
	}, tokenHash)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "register failed")
		return
	}
	_ = s.Store.AppendNodeEvent(r.Context(), n.ID, "enrolled", req.Name)
	_ = s.Store.AppendAudit(r.Context(), "agent:"+req.Name, "node.enroll", n.ID, req.AgentVersion)
	writeJSON(w, http.StatusCreated, enrollResp{NodeID: n.ID, NodeToken: nodeToken})
}

// handleAgentWS is implemented in agent_ws.go.
var _ = protocol.CurrentVersion
