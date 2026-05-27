package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func (s *Server) getSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if query == "" {
		http.Error(w, "missing query", http.StatusBadRequest)
		return
	}

	results, err := s.searchSvc.Search(r.Context(), query, limit)
	if err != nil {
		s.logger.Error("failed to search", "query", query, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Filter by lifecycle if provided
	lifecycle := r.URL.Query().Get("lifecycle")
	if lifecycle != "" {
		filtered := make([]*domain.SearchResult, 0)
		for _, res := range results {
			if res.Lifecycle == lifecycle {
				filtered = append(filtered, res)
			}
		}
		results = filtered
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *Server) getTimeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startsAtStr := r.URL.Query().Get("starts_at")
	endsAtStr := r.URL.Query().Get("ends_at")
	limitStr := r.URL.Query().Get("limit")
	kind := r.URL.Query().Get("kind")

	var startsAt, endsAt *time.Time
	if startsAtStr != "" {
		t, _ := time.Parse(time.RFC3339, startsAtStr)
		startsAt = &t
	}
	if endsAtStr != "" {
		t, _ := time.Parse(time.RFC3339, endsAtStr)
		endsAt = &t
	}

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	events, err := s.timeSvc.FetchTimeline(r.Context(), startsAt, endsAt, limit)
	if err != nil {
		s.logger.Error("failed to get timeline", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if kind != "" {
		filtered := make([]*domain.Event, 0)
		for _, e := range events {
			if string(e.Kind) == kind {
				filtered = append(filtered, e)
			}
		}
		events = filtered
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (s *Server) getActors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 100
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil {
			limit = l
		}
	}

	actors, err := s.idSvc.ListActors(r.Context(), limit)
	if err != nil {
		s.logger.Error("failed to list actors", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actors)
}

func (s *Server) getTeams(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 100
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil {
			limit = l
		}
	}

	teams, err := s.idSvc.ListTeams(r.Context(), limit)
	if err != nil {
		s.logger.Error("failed to list teams", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

func (s *Server) getZettel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/zettels/")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	z, err := s.syncSvc.GetZettel(r.Context(), id)
	if err != nil {
		s.logger.Error("failed to get zettel", "id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if z == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(z)
}

func (s *Server) getEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/events/")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	e, err := s.timeSvc.GetEvent(r.Context(), id)
	if err != nil {
		s.logger.Error("failed to get event", "id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if e == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.health)
	mux.Handle("/mcp", mcpcontroller.NewHandler(s.cfg, s.logger, s.ingestion, s.searchSvc, s.idSvc, s.timeSvc))

	mux.HandleFunc("/api/v1/sync/operations", s.syncOperations)
	mux.HandleFunc("/api/v1/sync/changes", s.syncChanges)
	mux.HandleFunc("/api/v1/sync/bootstrap", s.syncBootstrap)
	mux.HandleFunc("/api/v1/graph", s.getGraph)
	mux.HandleFunc("/api/v1/search", s.getSearch)
	mux.HandleFunc("/api/v1/timeline", s.getTimeline)
	mux.HandleFunc("/api/v1/actors", s.getActors)
	mux.HandleFunc("/api/v1/teams", s.getTeams)
	mux.HandleFunc("/api/v1/zettels/", s.getZettel)
	mux.HandleFunc("/api/v1/events/", s.getEvent)
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
