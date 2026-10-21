package admin

import (
	"net/http"

	"github.com/timsgarage/tapu/internal/session"
)

func (h *Handler) CurrentSession(
	writer http.ResponseWriter,
	request *http.Request,
) {
	user, userOK := session.UserFromContext(request.Context())
	found, sessionOK := session.FromContext(request.Context())

	if !userOK || !sessionOK {
		writeAPIError(writer, http.StatusUnauthorized, "Unauthorized")
		return
	}

	csrfCookie, err := request.Cookie("iam_csrf")
	if err != nil ||
		!h.sessions.ValidateCSRF(found, csrfCookie.Value) {
		writeAPIError(writer, http.StatusUnauthorized, "Unauthorized")
		return
	}

	writeJSON(writer, http.StatusOK, map[string]any{
		"user": userResponse{
			ID:        user.ID,
			Username:  user.Username,
			IsAdmin:   user.IsAdmin,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
		},
		"csrfToken": csrfCookie.Value,
	})
}
