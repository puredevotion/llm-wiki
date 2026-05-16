package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"llm-wiki/apps/backend/internal/config"
	mcpserver "llm-wiki/apps/backend/internal/mcp"
)

type Server struct {
	cfg    config.Config
	logger *slog.Logger
}

func NewServer(cfg config.Config, logger *slog.Logger) *Server {
	return &Server{cfg: cfg, logger: logger}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.Handle("/mcp", mcpserver.NewHandler(s.logger))
	return requestLogger(s.logger, mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("http request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
