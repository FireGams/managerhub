package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/managerhub/managerhub/controller/internal/db/store"
	"github.com/managerhub/managerhub/shared/protocol"
)

// multiJobReq targets multiple nodes or all online nodes.
type multiJobReq struct {
	Name     string            `json:"name"`
	Command  string            `json:"command"`
	Args     []string          `json:"args"`
	WorkDir  string            `json:"workdir"`
	Env      map[string]string `json:"env"`
	Timeout  int               `json:"timeout_sec"`
	NodeIDs  []string          `json:"node_ids"`  // specific nodes
	AllNodes bool              `json:"all_nodes"` // broadcast to all online
	Selector map[string]any    `json:"selector"`
}

// handleMultiJob creates and dispatches a job to multiple nodes.
func (s *Server) handleMultiJob(w http.ResponseWriter, r *http.Request) {
	var req multiJobReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Command == "" {
		writeErr(w, http.StatusBadRequest, "command is required")
		return
	}
	if req.Name == "" {
		req.Name = "multi-job"
	}
	if req.Timeout <= 0 {
		req.Timeout = 3600
	}

	// Collect target node IDs
	targets := req.NodeIDs
	if req.AllNodes || len(targets) == 0 {
		nodes, err := s.Store.ListNodes(r.Context())
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "db error")
			return
		}
		targets = nil
		for _, n := range nodes {
			if s.Hub.Online(n.ID) {
				targets = append(targets, n.ID)
			}
		}
	}
	if len(targets) == 0 {
		writeErr(w, http.StatusServiceUnavailable, "no online nodes")
		return
	}

	// Create and dispatch job to each node
	type jobInfo struct {
		JobID  string `json:"job_id"`
		NodeID string `json:"node_id"`
		Name   string `json:"name"`
	}
	var created []jobInfo
	for _, nodeID := range targets {
		nid := nodeID
		j, err := s.Store.CreateJob(r.Context(), store.Job{
			ID: uuid.NewString(), Name: req.Name, Type: "shell",
			Command: req.Command, Args: req.Args, WorkDir: req.WorkDir,
			Env: req.Env, TimeoutSec: req.Timeout, NodeID: &nid,
		})
		if err != nil {
			continue
		}
		s.assignJob(j)
		created = append(created, jobInfo{JobID: j.ID, NodeID: nid, Name: req.Name})
	}
	writeJSON(w, http.StatusOK, map[string]any{"jobs": created, "count": len(created)})
}

// runnerInstallReq requests runner installation on one or more nodes.
type runnerInstallReq struct {
	NodeIDs  []string `json:"node_ids"`
	AllNodes bool     `json:"all_nodes"`
	RepoURL  string   `json:"repo_url"`
	Token    string   `json:"token"`
	Name     string   `json:"name"`
	Labels   string   `json:"labels"`
	WorkDir  string   `json:"workdir"`
}

// handleRunnerInstall pushes a runner installation to selected nodes.
func (s *Server) handleRunnerInstall(w http.ResponseWriter, r *http.Request) {
	var req runnerInstallReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.RepoURL == "" || req.Token == "" {
		writeErr(w, http.StatusBadRequest, "repo_url and token are required")
		return
	}

	targets := req.NodeIDs
	if req.AllNodes || len(targets) == 0 {
		nodes, _ := s.Store.ListNodes(r.Context())
		targets = nil
		for _, n := range nodes {
			if s.Hub.Online(n.ID) {
				targets = append(targets, n.ID)
			}
		}
	}

	results := make([]map[string]string, 0)
	for _, nodeID := range targets {
		name := req.Name
		if name == "" {
			name = "runner-" + nodeID[:8]
		}
		env, _ := protocol.NewEnvelope(protocol.TypeRunnerInstall, "rinst-"+uuid.NewString(), time.Now().Unix(), nodeID,
			protocol.RunnerInstall{
				ReqID: uuid.NewString(), RepoURL: req.RepoURL, Token: req.Token,
				Name: name, Labels: req.Labels, WorkDir: req.WorkDir,
			})
		if err := s.Hub.Send(nodeID, env); err == nil {
			results = append(results, map[string]string{"node_id": nodeID, "status": "sent"})
		} else {
			results = append(results, map[string]string{"node_id": nodeID, "status": "error", "error": err.Error()})
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"sent": results, "count": len(results)})
}
