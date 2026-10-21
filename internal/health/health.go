package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type DatabasePinger interface {
	PingContext(ctx context.Context) error
}

type Handler struct {
	database DatabasePinger
}

type response struct {
	Status string `json:"status"`
}

func NewHandler(database DatabasePinger) *Handler {
	return &Handler{
		database: database,
	}
}

func (h *Handler) Live(
	writer http.ResponseWriter,
	_ *http.Request,
) {
	writeJSON(writer, http.StatusOK, response{
		Status: "ok",
	})
}

func (h *Handler) Ready(
	writer http.ResponseWriter,
	request *http.Request,
) {
	ctx, cancel := context.WithTimeout(request.Context(), 2*time.Second)
	defer cancel()

	if err := h.database.PingContext(ctx); err != nil {
		writeJSON(writer, http.StatusServiceUnavailable, response{
			Status: "unavailable",
		})
		return
	}

	writeJSON(writer, http.StatusOK, response{
		Status: "ok",
	})
}

func writeJSON(
	writer http.ResponseWriter,
	status int,
	value any,
) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)

	_ = json.NewEncoder(writer).Encode(value)
}
