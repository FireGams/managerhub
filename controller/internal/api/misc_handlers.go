package api

import (
	"net"
	"net/http"
)

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
	url := s.Cfg.PublicURL
	if url == "" || url == "http://localhost:8080" {
		url = detectLocalIP()
	}
	writeJSON(w, http.StatusOK, map[string]string{"public_url": url})
}

// detectLocalIP finds the machine's primary LAN IP automatically.
func detectLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "http://localhost:8080"
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return "http://" + ipNet.IP.String() + ":8080"
		}
	}
	return "http://localhost:8080"
}
