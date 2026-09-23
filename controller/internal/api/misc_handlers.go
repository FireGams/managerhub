package api

import "net/http"

func (s *Server) handleListAudit(w http.ResponseWriter, r *http.Request) {
	list, err := s.Store.ListAudit(r.Context(), 200)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

var schedulerReload = func() {}

// SetSchedulerReload wires the cron engine refresh callback.
func SetSchedulerReload(fn func()) { schedulerReload = fn }

func (s *Server) reloadScheduler() { schedulerReload() }

func (s *Server) handlePublicConfig(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"public_url": s.Cfg.PublicURL})
}
