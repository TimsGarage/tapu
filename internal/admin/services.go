package admin

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

func (h *Handler) ListServices(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var services []Service

	if err := h.database.WithContext(request.Context()).
		Order("name ASC").
		Find(&services).
		Error; err != nil {
		h.internalError(writer, request, err)
		return
	}

	writeJSON(writer, http.StatusOK, services)
}

func (h *Handler) CreateService(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var input struct {
		Key      string `json:"key"`
		Name     string `json:"name"`
		Audience string `json:"audience"`
	}

	if err := decodeJSON(writer, request, &input); err != nil {
		writeAPIError(writer, http.StatusBadRequest, "Invalid request")
		return
	}

	input.Key = strings.ToLower(strings.TrimSpace(input.Key))
	input.Name = strings.TrimSpace(input.Name)
	input.Audience = strings.TrimSpace(input.Audience)

	if err := validateServiceKey(input.Key); err != nil {
		writeAPIError(writer, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateName(input.Name); err != nil {
		writeAPIError(writer, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateAudience(input.Audience); err != nil {
		writeAPIError(writer, http.StatusBadRequest, err.Error())
		return
	}

	now := time.Now().UTC()
	service := Service{
		ID:        uuid.New(),
		Key:       input.Key,
		Name:      input.Name,
		Audience:  input.Audience,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.database.WithContext(request.Context()).
		Create(&service).
		Error; err != nil {
		if isUniqueViolation(err) {
			writeAPIError(
				writer,
				http.StatusConflict,
				"Service key or audience already exists",
			)
			return
		}

		h.internalError(writer, request, err)
		return
	}

	h.recordAudit(
		request,
		"service.created",
		"service",
		&service.ID,
		map[string]any{"key": service.Key},
	)

	writeJSON(writer, http.StatusCreated, service)
}

func (h *Handler) UpdateService(
	writer http.ResponseWriter,
	request *http.Request,
) {
	serviceID, err := parseUUID(chi.URLParam(request, "serviceID"))
	if err != nil {
		writeAPIError(writer, http.StatusBadRequest, "Invalid service ID")
		return
	}

	var input struct {
		Name     *string `json:"name"`
		Audience *string `json:"audience"`
		Enabled  *bool   `json:"enabled"`
	}

	if err := decodeJSON(writer, request, &input); err != nil {
		writeAPIError(writer, http.StatusBadRequest, "Invalid request")
		return
	}

	updates := map[string]any{"updated_at": time.Now().UTC()}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if err := validateName(name); err != nil {
			writeAPIError(writer, http.StatusBadRequest, err.Error())
			return
		}
		updates["name"] = name
	}

	if input.Audience != nil {
		audience := strings.TrimSpace(*input.Audience)
		if err := validateAudience(audience); err != nil {
			writeAPIError(writer, http.StatusBadRequest, err.Error())
			return
		}
		updates["audience"] = audience
	}

	if input.Enabled != nil {
		updates["enabled"] = *input.Enabled
	}

	result := h.database.WithContext(request.Context()).
		Model(&Service{}).
		Where("id = ?", serviceID).
		Updates(updates)

	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			writeAPIError(
				writer,
				http.StatusConflict,
				"Audience already exists",
			)
			return
		}

		h.internalError(writer, request, result.Error)
		return
	}

	if result.RowsAffected == 0 {
		writeAPIError(writer, http.StatusNotFound, "Service not found")
		return
	}

	var service Service
	if err := h.database.WithContext(request.Context()).
		First(&service, "id = ?", serviceID).
		Error; err != nil {
		h.internalError(writer, request, err)
		return
	}

	h.recordAudit(
		request,
		"service.updated",
		"service",
		&serviceID,
		nil,
	)

	writeJSON(writer, http.StatusOK, service)
}

func (h *Handler) ListUserServices(
	writer http.ResponseWriter,
	request *http.Request,
) {
	userID, err := parseUUID(chi.URLParam(request, "userID"))
	if err != nil {
		writeAPIError(writer, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var services []Service

	err = h.database.WithContext(request.Context()).
		Table("services").
		Joins(
			"JOIN user_service_access usa ON usa.service_id = services.id",
		).
		Where("usa.user_id = ?", userID).
		Order("services.name ASC").
		Find(&services).
		Error

	if err != nil {
		h.internalError(writer, request, err)
		return
	}

	writeJSON(writer, http.StatusOK, services)
}

func (h *Handler) GrantUserService(
	writer http.ResponseWriter,
	request *http.Request,
) {
	userID, userErr := parseUUID(chi.URLParam(request, "userID"))
	serviceID, serviceErr := parseUUID(
		chi.URLParam(request, "serviceID"),
	)

	if userErr != nil || serviceErr != nil {
		writeAPIError(writer, http.StatusBadRequest, "Invalid ID")
		return
	}

	actor, _ := currentUser(request)

	access := UserServiceAccess{
		UserID:    userID,
		ServiceID: serviceID,
		CreatedBy: &actor.ID,
		CreatedAt: time.Now().UTC(),
	}

	result := h.database.WithContext(request.Context()).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&access)

	if result.Error != nil {
		if isForeignKeyViolation(result.Error) {
			writeAPIError(
				writer,
				http.StatusNotFound,
				"User or service not found",
			)
			return
		}

		h.internalError(writer, request, result.Error)
		return
	}

	h.recordAudit(
		request,
		"user.service_granted",
		"user",
		&userID,
		map[string]any{"serviceId": serviceID},
	)

	writeJSON(writer, http.StatusNoContent, nil)
}

func (h *Handler) RevokeUserService(
	writer http.ResponseWriter,
	request *http.Request,
) {
	userID, userErr := parseUUID(chi.URLParam(request, "userID"))
	serviceID, serviceErr := parseUUID(
		chi.URLParam(request, "serviceID"),
	)

	if userErr != nil || serviceErr != nil {
		writeAPIError(writer, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := h.database.WithContext(request.Context()).
		Delete(
			&UserServiceAccess{},
			"user_id = ? AND service_id = ?",
			userID,
			serviceID,
		).
		Error; err != nil {
		h.internalError(writer, request, err)
		return
	}

	h.recordAudit(
		request,
		"user.service_revoked",
		"user",
		&userID,
		map[string]any{"serviceId": serviceID},
	)

	writeJSON(writer, http.StatusNoContent, nil)
}
