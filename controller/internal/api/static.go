package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// serveStatic serves the SPA from web/dist, falling back to index.html.
func (s *Server) serveStatic(w http.ResponseWriter, r *http.Request) {
	root := os.Getenv("MH_WEB_ROOT")
	if root == "" {
		root = "web/dist"
	}
	p := filepath.Join(root, filepath.Clean("/"+r.URL.Path))
	if st, err := os.Stat(p); err == nil && !st.IsDir() {
		http.ServeFile(w, r, p)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	http.ServeFile(w, r, filepath.Join(root, "index.html"))
}
