package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/managerhub/managerhub/controller/internal/db/store"
	"github.com/robfig/cron/v3"
)

type createSchedReq struct {
	Name       string            `json:"name"`
	CronExpr   string            `json:"cron_expr"`
	Enabled    *bool             `json:"enabled"`
	JobName    string            `json:"job_name"`
	JobType    string            `json:"job_type"`
	Command    string            `json:"command"`
	Args       []string          `json:"args"`
	WorkDir    string            `json:"workdir"`
	Env        map[string]string `json:"env"`
	TimeoutSec int               `json:"timeout_sec"`
	NodeID     string            `json:"node_id"`
	Selector   map[string]any    `json:"selector"`
}

func (s *Server) handleCreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req createSchedReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Name == "" || req.CronExpr == "" || req.Command == "" {
		writeErr(w, http.StatusBadRequest, "name, cron_expr and command are required")
		return
	}
	if _, err := cron.ParseStandard(req.CronExpr); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid cron expression: "+err.Error())
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	if req.JobName == "" {
		req.JobName = req.Name
	}
	if req.JobType == "" {
		req.JobType = "shell"
	}
	if req.TimeoutSec <= 0 {
		req.TimeoutSec = 3600
	}
	var nodeID *string
	if req.NodeID != "" {
		nodeID = &req.NodeID
	}
	sc, err := s.Store.CreateSchedule(r.Context(), store.Schedule{
		ID: uuid.NewString(), Name: req.Name, CronExpr: req.CronExpr, Enabled: enabled,
		JobName: req.JobName, JobType: req.JobType, Command: req.Command, Args: req.Args,
		WorkDir: req.WorkDir, Env: req.Env, TimeoutSec: req.TimeoutSec, NodeID: nodeID,
		Selector: req.Selector,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "create schedule failed")
		return
	}
	s.reloadScheduler()
	writeJSON(w, http.StatusCreated, sc)
}

func (s *Server) handleListSchedules(w http.ResponseWriter, r *http.Request) {
	list, err := s.Store.ListSchedules(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

type patchSchedReq struct {
	Enabled *bool `json:"enabled"`
}

func (s *Server) handlePatchSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req patchSchedReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Enabled == nil {
		writeErr(w, http.StatusBadRequest, "enabled is required")
		return
	}
	if err := s.Store.SetScheduleEnabled(r.Context(), id, *req.Enabled); err != nil {
		writeErr(w, http.StatusInternalServerError, "update failed")
		return
	}
	s.reloadScheduler()
	writeJSON(w, http.StatusOK, map[string]bool{"enabled": *req.Enabled})
}
