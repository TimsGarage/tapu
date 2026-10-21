package admin

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/timsgarage/tapu/internal/audit"
	"github.com/timsgarage/tapu/internal/identity"
	"github.com/timsgarage/tapu/internal/session"
	"gorm.io/gorm"
)

const maximumJSONBodySize = 64 * 1024

type Handler struct {
	logger    *slog.Logger
	database  *gorm.DB
	passwords *identity.PasswordHasher
	sessions  *session.Manager
	audit     *audit.Repository
}

func NewHandler(
	logger *slog.Logger,
	database *gorm.DB,
	passwords *identity.PasswordHasher,
	sessions *session.Manager,
	auditRepository *audit.Repository,
) *Handler {
	return &Handler{
		logger:    logger,
		database:  database,
		passwords: passwords,
		sessions:  sessions,
		audit:     auditRepository,
	}
}

func decodeJSON(
	writer http.ResponseWriter,
	request *http.Request,
	target any,
) error {
	request.Body = http.MaxBytesReader(
		writer,
		request.Body,
		maximumJSONBodySize,
	)

	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON object")
	}

	return nil
}

func writeJSON(
	writer http.ResponseWriter,
	status int,
	value any,
) {
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(status)

	if value != nil {
		_ = json.NewEncoder(writer).Encode(value)
	}
}

func writeAPIError(
	writer http.ResponseWriter,
	status int,
	message string,
) {
	writeJSON(writer, status, map[string]string{
		"error": message,
	})
}

func parseUUID(value string) (uuid.UUID, error) {
	return uuid.Parse(strings.TrimSpace(value))
}

func currentUser(request *http.Request) (*identity.User, bool) {
	return session.UserFromContext(request.Context())
}

func (h *Handler) internalError(
	writer http.ResponseWriter,
	request *http.Request,
	err error,
) {
	h.logger.Error(
		"admin API request failed",
		"error",
		err,
		"request_id",
		middleware.GetReqID(request.Context()),
	)

	writeAPIError(
		writer,
		http.StatusInternalServerError,
		"Internal server error",
	)
}

func (h *Handler) recordAudit(
	request *http.Request,
	action string,
	targetType string,
	targetID *uuid.UUID,
	metadata map[string]any,
) {
	actor, ok := currentUser(request)
	if !ok {
		return
	}

	event := audit.NewEvent(audit.ActorUser, action)
	event.ActorID = &actor.ID
	event.TargetType = &targetType
	event.TargetID = targetID

	requestID := middleware.GetReqID(request.Context())
	event.RequestID = &requestID

	if err := h.audit.Record(
		request.Context(),
		event,
		metadata,
	); err != nil {
		h.logger.Error(
			"record administrative audit event",
			"error",
			err,
			"action",
			action,
		)
	}
}
