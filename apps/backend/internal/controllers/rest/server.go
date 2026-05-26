package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"llm-wiki/apps/backend/internal/config"
	mcpcontroller "llm-wiki/apps/backend/internal/controllers/mcp"
	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/services"
)

type Server struct {
	cfg       config.Config
	logger    *slog.Logger
	ingestion *services.IngestionService
	searchSvc *services.SearchService
	syncSvc   *services.SyncService
	idSvc     *services.IdentityService
	timeSvc   *services.TimelineService
	graphSvc  *services.GraphService
}

func NewServer(cfg config.Config, logger *slog.Logger, ingestion *services.IngestionService, searchSvc *services.SearchService, syncSvc *services.SyncService, idSvc *services.IdentityService, timeSvc *services.TimelineService, graphSvc *services.GraphService) *Server {
	return &Server{cfg: cfg, logger: logger, ingestion: ingestion, searchSvc: searchSvc, syncSvc: syncSvc, idSvc: idSvc, timeSvc: timeSvc, graphSvc: graphSvc}
}

func (s *Server) getGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := s.graphSvc.GetGraph(r.Context())
	if err != nil {
		s.logger.Error("failed to get graph", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.health)
	mux.Handle("/mcp", mcpcontroller.NewHandler(s.cfg, s.logger, s.ingestion, s.searchSvc, s.idSvc, s.timeSvc))

	mux.HandleFunc("/api/v1/sync/operations", s.syncOperations)
	mux.HandleFunc("/api/v1/sync/changes", s.syncChanges)
	mux.HandleFunc("/api/v1/sync/bootstrap", s.syncBootstrap)
	mux.HandleFunc("/api/v1/graph", s.getGraph)
	mux.HandleFunc("/api//", s.notFound)

	return requestLogger(s.logger, mux)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) syncOperations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var batch domain.SyncBatch
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	results, err := s.syncSvc.ProcessBatch(r.Context(), batch)
	if err != nil {
		s.logger.Error("failed to process sync batch", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"results": results})
}

func (s *Server) syncChanges(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cursor := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	changes, err := s.syncSvc.FetchChanges(r.Context(), cursor, limit)
	if err != nil {
		s.logger.Error("failed to fetch changes", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"changes": changes})
}

func (s *Server) syncBootstrap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Not implemented for now, but reserved for full database dump
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not found", http.StatusNotFound)
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("request", "method", r.Method, "url", r.URL.String())
		next.ServeHTTP(w, r)
	})
}
