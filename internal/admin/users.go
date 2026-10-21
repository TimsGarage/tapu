package admin

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/timsgarage/tapu/internal/identity"
	"gorm.io/gorm"
)

type userResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IsAdmin   bool      `json:"isAdmin"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

func userToResponse(user identity.User) userResponse {
	return userResponse{
		ID:        user.ID,
		Username:  user.Username,
		IsAdmin:   user.IsAdmin,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
	}
}

func (h *Handler) ListUsers(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var users []identity.User

	if err := h.database.WithContext(request.Context()).
		Order("username ASC").
		Find(&users).
		Error; err != nil {
		h.internalError(writer, request, err)
		return
	}

	response := make([]userResponse, 0, len(users))
	for _, user := range users {
		response = append(response, userToResponse(user))
	}

	writeJSON(writer, http.StatusOK, response)
}

func (h *Handler) CreateUser(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
		IsAdmin  bool   `json:"isAdmin"`
	}

	if err := decodeJSON(writer, request, &input); err != nil {
		writeAPIError(writer, http.StatusBadRequest, "Invalid request")
		return
	}

	input.Username = identity.NormalizeUsername(input.Username)

	if err := identity.ValidateUsername(input.Username); err != nil {
		writeAPIError(writer, http.StatusBadRequest, err.Error())
		return
	}

	passwordHash, err := h.passwords.Hash(input.Password)
	if err != nil {
		writeAPIError(writer, http.StatusBadRequest, err.Error())
		return
	}

	now := time.Now().UTC()
	user := identity.User{
		ID:                uuid.New(),
		Username:          input.Username,
		PasswordHash:      passwordHash,
		IsAdmin:           input.IsAdmin,
		Status:            identity.StatusActive,
		PasswordChangedAt: now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := h.database.WithContext(request.Context()).
		Create(&user).
		Error; err != nil {
		if isUniqueViolation(err) {
			writeAPIError(
				writer,
				http.StatusConflict,
				"Username already exists",
			)
			return
		}

		h.internalError(writer, request, err)
		return
	}

	h.recordAudit(
		request,
		"user.created",
		"user",
		&user.ID,
		map[string]any{
			"username": user.Username,
			"isAdmin":  user.IsAdmin,
		},
	)

	writeJSON(writer, http.StatusCreated, userToResponse(user))
}

func (h *Handler) UpdateUser(
	writer http.ResponseWriter,
	request *http.Request,
) {
	userID, err := parseUUID(chi.URLParam(request, "userID"))
	if err != nil {
		writeAPIError(writer, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var input struct {
		Username *string `json:"username"`
		IsAdmin  *bool   `json:"isAdmin"`
		Status   *string `json:"status"`
	}

	if err := decodeJSON(writer, request, &input); err != nil {
		writeAPIError(writer, http.StatusBadRequest, "Invalid request")
		return
	}

	var user identity.User

	if err := h.database.WithContext(request.Context()).
		First(&user, "id = ?", userID).
		Error; errors.Is(err, gorm.ErrRecordNotFound) {
		writeAPIError(writer, http.StatusNotFound, "User not found")
		return
	} else if err != nil {
		h.internalError(writer, request, err)
		return
	}

	updates := map[string]any{
		"updated_at": time.Now().UTC(),
	}

	if input.Username != nil {
		username := identity.NormalizeUsername(*input.Username)

		if err := identity.ValidateUsername(username); err != nil {
			writeAPIError(writer, http.StatusBadRequest, err.Error())
			return
		}

		updates["username"] = username
	}

	if input.IsAdmin != nil {
		updates["is_admin"] = *input.IsAdmin
	}

	if input.Status != nil {
		status := strings.TrimSpace(*input.Status)

		if status != identity.StatusActive &&
			status != identity.StatusDisabled {
			writeAPIError(writer, http.StatusBadRequest, "Invalid status")
			return
		}

		updates["status"] = status
	}

	if err := h.database.WithContext(request.Context()).
		Model(&identity.User{}).
		Where("id = ?", userID).
		Updates(updates).
		Error; err != nil {
		if isUniqueViolation(err) {
			writeAPIError(
				writer,
				http.StatusConflict,
				"Username already exists",
			)
			return
		}

		h.internalError(writer, request, err)
		return
	}

	if input.Status != nil && *input.Status == identity.StatusDisabled {
		if err := h.sessions.DeleteAllForUser(
			request.Context(),
			userID,
		); err != nil {
			h.internalError(writer, request, err)
			return
		}
	}

	if err := h.database.WithContext(request.Context()).
		First(&user, "id = ?", userID).
		Error; err != nil {
		h.internalError(writer, request, err)
		return
	}

	h.recordAudit(
		request,
		"user.updated",
		"user",
		&userID,
		nil,
	)

	writeJSON(writer, http.StatusOK, userToResponse(user))
}

func (h *Handler) ResetPassword(
	writer http.ResponseWriter,
	request *http.Request,
) {
	userID, err := parseUUID(chi.URLParam(request, "userID"))
	if err != nil {
		writeAPIError(writer, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var input struct {
		Password string `json:"password"`
	}

	if err := decodeJSON(writer, request, &input); err != nil {
		writeAPIError(writer, http.StatusBadRequest, "Invalid request")
		return
	}

	passwordHash, err := h.passwords.Hash(input.Password)
	if err != nil {
		writeAPIError(writer, http.StatusBadRequest, err.Error())
		return
	}

	result := h.database.WithContext(request.Context()).
		Model(&identity.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"password_hash":       passwordHash,
			"password_changed_at": time.Now().UTC(),
			"updated_at":          time.Now().UTC(),
		})

	if result.Error != nil {
		h.internalError(writer, request, result.Error)
		return
	}

	if result.RowsAffected == 0 {
		writeAPIError(writer, http.StatusNotFound, "User not found")
		return
	}

	if err := h.sessions.DeleteAllForUser(
		request.Context(),
		userID,
	); err != nil {
		h.internalError(writer, request, err)
		return
	}

	h.recordAudit(
		request,
		"user.password_reset",
		"user",
		&userID,
		nil,
	)

	writeJSON(writer, http.StatusNoContent, nil)
}
