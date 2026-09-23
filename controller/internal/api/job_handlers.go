package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/managerhub/managerhub/controller/internal/db/store"
	"github.com/managerhub/managerhub/shared/protocol"
)

type createJobReq struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	NodeID     string            `json:"node_id"`
	Command    string            `json:"command"`
	Args       []string          `json:"args"`
	WorkDir    string            `json:"workdir"`
	Env        map[string]string `json:"env"`
	TimeoutSec int               `json:"timeout_sec"`
	Selector   map[string]any    `json:"selector"`
}

func (s *Server) handleCreateJob(w http.ResponseWriter, r *http.Request) {
	var req createJobReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Name == "" || req.Command == "" {
		writeErr(w, http.StatusBadRequest, "name and command are required")
		return
	}
	if req.Type == "" {
		req.Type = "shell"
	}
	if req.TimeoutSec <= 0 {
		req.TimeoutSec = 3600
	}
	var nodeID *string
	if req.NodeID != "" {
		nodeID = &req.NodeID
	} else {
		id, err := s.PickNode(r.Context(), req.Selector)
		if err != nil {
			writeErr(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		nodeID = &id
	}
	j, err := s.Store.CreateJob(r.Context(), store.Job{
		ID: uuid.NewString(), Name: req.Name, Type: req.Type, NodeID: nodeID,
		Command: req.Command, Args: req.Args, WorkDir: req.WorkDir, Env: req.Env,
		TimeoutSec: req.TimeoutSec,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "create job failed")
		return
	}
	s.assignJob(j)
	writeJSON(w, http.StatusCreated, j)
}

// assignJob pushes a job to its target node over the hub.
func (s *Server) assignJob(j store.Job) {
	if j.NodeID == nil {
		return
	}
	payload := protocol.JobAssign{
		JobID: j.ID, Name: j.Name, Type: j.Type, Command: j.Command,
		Args: j.Args, WorkDir: j.WorkDir, Env: j.Env, TimeoutSec: j.TimeoutSec,
	}
	env, err := protocol.NewEnvelope(protocol.TypeJobAssign, "assign-"+j.ID, nowUnix(), *j.NodeID, payload)
	if err != nil {
		return
	}
	if err := s.Hub.Send(*j.NodeID, env); err == nil {
		_ = s.Store.UpdateJobStatus(context.Background(), j.ID, protocol.JobAssigned, nil, "")
	}
}

func (s *Server) handleListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := s.Store.ListJobs(r.Context(), 100)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, jobs)
}

func (s *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	j, err := s.Store.GetJob(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, "job not found")
		return
	}
	writeJSON(w, http.StatusOK, j)
}

func (s *Server) handleCancelJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	j, err := s.Store.GetJob(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "job not found")
		return
	}
	if j.NodeID != nil {
		env, _ := protocol.NewEnvelope(protocol.TypeJobCancel, "cancel-"+id, nowUnix(), *j.NodeID,
			protocol.JobCancel{JobID: id})
		_ = s.Hub.Send(*j.NodeID, env)
	}
	code := -1
	_ = s.Store.UpdateJobStatus(r.Context(), id, protocol.JobCancelled, &code, "cancelled by user")
	writeJSON(w, http.StatusOK, map[string]string{"status": protocol.JobCancelled})
}
