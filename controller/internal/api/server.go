// Package api exposes the REST + WebSocket surface of the controller.
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/managerhub/managerhub/controller/internal/auth"
	"github.com/managerhub/managerhub/controller/internal/config"
	"github.com/managerhub/managerhub/controller/internal/db/store"
	"github.com/managerhub/managerhub/controller/internal/hub"
	"github.com/managerhub/managerhub/shared/protocol"
)

// Server bundles every dependency of the HTTP layer.
type Server struct {
	Cfg   config.Config
	Store *store.Store
	Hub   *hub.Hub
	Log   *slog.Logger

	uiMu    sync.RWMutex
	uiClients map[*broadcaster]bool
}

// Router builds the chi mux.
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
	}))

	r.Get("/healthz", s.handleHealth)

	// WebSocket for web UI (terminal, live updates)
	r.Get("/api/v1/ws", s.handleUIWS)

	// SPA static files
	r.NotFound(s.serveStatic)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/login", s.handleLogin)
		r.Post("/auth/bootstrap", s.handleBootstrap)

		// Agent enrollment (one-time token) and WebSocket.
		r.Post("/agent/enroll", s.handleEnroll)
		r.Get("/agent/ws", s.handleAgentWS)

		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(s.Cfg.JWTSecret))
			r.Get("/me", s.handleMe)

			r.Get("/nodes", s.handleListNodes)
			r.Get("/nodes/{id}", s.handleGetNode)
			r.Get("/nodes/{id}/metrics", s.handleNodeMetrics)

			r.Get("/jobs", s.handleListJobs)
			r.Get("/jobs/{id}", s.handleGetJob)
			r.With(auth.RequireRole(protocol.RoleOperator)).Post("/jobs", s.handleCreateJob)
			r.With(auth.RequireRole(protocol.RoleOperator)).Post("/jobs/{id}/cancel", s.handleCancelJob)

			r.Get("/schedules", s.handleListSchedules)
			r.With(auth.RequireRole(protocol.RoleOperator)).Post("/schedules", s.handleCreateSchedule)
			r.With(auth.RequireRole(protocol.RoleOperator)).Patch("/schedules/{id}", s.handlePatchSchedule)

			r.Get("/audit", s.handleListAudit)
			r.With(auth.RequireRole(protocol.RoleAdmin)).Post("/enroll-tokens", s.handleCreateEnrollToken)
			r.Get("/enroll-tokens", s.handleListEnrollTokens)

			r.Post("/ai/chat", s.handleAIChat)
			r.Get("/ai/config", s.handleAIConfig)
			r.With(auth.RequireRole(protocol.RoleAdmin)).Post("/ai/config", s.handleAISetConfig)
			r.Get("/ai/models", s.handleAIModels)
		})
	})
	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
