package httpapi

import (
	"encoding/json"
	"net/http"

	"log/slog"

	"omnibase/internal/flow"
	"omnibase/internal/logging"
	"omnibase/internal/schema"
)

type Handler struct {
	Flow   flow.Flow
	Logger *slog.Logger
}

func NewHandler(flow flow.Flow, logger *slog.Logger) *Handler {
	return &Handler{Flow: flow, Logger: logger}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req schema.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	ctx := r.Context()
	logger, ctx := logging.WithRequest(ctx, h.Logger, req.RequestID, req.TraceID, "http")
	result, err := h.Flow.Execute(ctx, req)
	if err != nil {
		logger.Error("flow execution failed", "error", err.Error())
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("content-type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		logger.Error("encode response failed", "error", err.Error())
		return
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
