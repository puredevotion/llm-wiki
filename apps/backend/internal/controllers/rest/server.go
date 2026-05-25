package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

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
}

func NewServer(cfg config.Config, logger *slog.Logger, ingestion *services.IngestionService, searchSvc *services.SearchService, syncSvc *services.SyncService, idSvc *services.IdentityService, timeSvc *services.TimelineService) *Server {
	return &Server{cfg: cfg, logger: logger, ingestion: ingestion, searchSvc: searchSvc, syncSvc: syncSvc, idSvc: idSvc, timeSvc: timeSvc}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.health)
	mux.Handle("/mcp", mcpcontroller.NewHandler(s.cfg, s.logger, s.ingestion, s.searchSvc, s.idSvc, s.timeSvc))

	mux.HandleFunc("/api/v1/sync/operations", s.syncOperations)
	mux.HandleFunc("/api/v1/sync/changes", s.syncChanges)
	mux.HandleFunc("/api/v1/sync/bootstrap", s.syncBootstrap)
	mux.HandleFunc("/api/", s.notFound)
	return requestLogger(s.logger, mux)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}
	writeJSON(w, http.StatusOK, successEnvelope{Data: map[string]string{"status": "ok"}})
}

func (s *Server) syncOperations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	var batch domain.SyncBatch
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Failed to decode sync batch.")
		return
	}

	results, err := s.syncSvc.ProcessBatch(r.Context(), batch)
	if err != nil {
		s.logger.Error("failed to process sync batch", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to process operations.")
		return
	}

	writeJSON(w, http.StatusOK, successEnvelope{Data: results})
}

func (s *Server) syncChanges(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	cursor := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	changes, err := s.syncSvc.FetchChanges(r.Context(), cursor, limit)
	if err != nil {
		s.logger.Error("failed to fetch sync changes", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch changes.")
		return
	}

	writeJSON(w, http.StatusOK, successEnvelope{Data: changes})
}

func (s *Server) syncBootstrap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}
	writeError(w, http.StatusNotImplemented, "not_implemented", "Sync bootstrap is not implemented yet.")
}

func (s *Server) notFound(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, "not_found", "The requested API endpoint was not found.")
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("http request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

type successEnvelope struct {
	Data any `json:"data"`
}

type errorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeMethodNotAllowed(w http.ResponseWriter, allowedMethods ...string) {
	w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "The requested HTTP method is not allowed for this endpoint.")
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, errorEnvelope{Error: apiError{Code: code, Message: message}})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
