package todos

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/timsgarage/tapu/sdk/go/authsdk"
)

const maximumRequestSize = 16 * 1024

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{
		store: store,
	}
}

func (h *Handler) List(
	writer http.ResponseWriter,
	request *http.Request,
) {
	identity, ok := authsdk.IdentityFromContext(
		request.Context(),
	)
	if !ok {
		writeError(
			writer,
			http.StatusUnauthorized,
			"unauthenticated",
		)
		return
	}

	writeJSON(
		writer,
		http.StatusOK,
		h.store.List(identity.Subject),
	)
}

func (h *Handler) Create(
	writer http.ResponseWriter,
	request *http.Request,
) {
	identity, ok := authsdk.IdentityFromContext(
		request.Context(),
	)
	if !ok {
		writeError(
			writer,
			http.StatusUnauthorized,
			"unauthenticated",
		)
		return
	}

	var input struct {
		Title string `json:"title"`
	}

	if err := decodeJSON(
		writer,
		request,
		&input,
	); err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
		)
		return
	}

	input.Title = strings.TrimSpace(input.Title)

	if len([]rune(input.Title)) < 1 ||
		len([]rune(input.Title)) > 200 {
		writeError(
			writer,
			http.StatusBadRequest,
			"title must contain between 1 and 200 characters",
		)
		return
	}

	todo := h.store.Create(
		identity.Subject,
		input.Title,
	)

	writeJSON(
		writer,
		http.StatusCreated,
		todo,
	)
}

func (h *Handler) Update(
	writer http.ResponseWriter,
	request *http.Request,
) {
	identity, ok := authsdk.IdentityFromContext(
		request.Context(),
	)
	if !ok {
		writeError(
			writer,
			http.StatusUnauthorized,
			"unauthenticated",
		)
		return
	}

	var input struct {
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
	}

	if err := decodeJSON(
		writer,
		request,
		&input,
	); err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
		)
		return
	}

	input.Title = strings.TrimSpace(input.Title)

	if len([]rune(input.Title)) < 1 ||
		len([]rune(input.Title)) > 200 {
		writeError(
			writer,
			http.StatusBadRequest,
			"title must contain between 1 and 200 characters",
		)
		return
	}

	todo, err := h.store.Update(
		identity.Subject,
		chi.URLParam(request, "todoID"),
		input.Title,
		input.Completed,
	)
	if errors.Is(err, ErrNotFound) {
		writeError(
			writer,
			http.StatusNotFound,
			"todo_not_found",
		)
		return
	}

	if err != nil {
		writeError(
			writer,
			http.StatusInternalServerError,
			"internal_error",
		)
		return
	}

	writeJSON(writer, http.StatusOK, todo)
}

func (h *Handler) Delete(
	writer http.ResponseWriter,
	request *http.Request,
) {
	identity, ok := authsdk.IdentityFromContext(
		request.Context(),
	)
	if !ok {
		writeError(
			writer,
			http.StatusUnauthorized,
			"unauthenticated",
		)
		return
	}

	err := h.store.Delete(
		identity.Subject,
		chi.URLParam(request, "todoID"),
	)

	if errors.Is(err, ErrNotFound) {
		writeError(
			writer,
			http.StatusNotFound,
			"todo_not_found",
		)
		return
	}

	if err != nil {
		writeError(
			writer,
			http.StatusInternalServerError,
			"internal_error",
		)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func decodeJSON(
	writer http.ResponseWriter,
	request *http.Request,
	target any,
) error {
	request.Body = http.MaxBytesReader(
		writer,
		request.Body,
		maximumRequestSize,
	)

	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(
		err,
		io.EOF,
	) {
		return errors.New(
			"request must contain one JSON object",
		)
	}

	return nil
}

func writeError(
	writer http.ResponseWriter,
	status int,
	message string,
) {
	writeJSON(
		writer,
		status,
		map[string]string{
			"error": message,
		},
	)
}

func writeJSON(
	writer http.ResponseWriter,
	status int,
	value any,
) {
	writer.Header().Set(
		"Content-Type",
		"application/json",
	)
	writer.Header().Set(
		"Cache-Control",
		"no-store",
	)
	writer.WriteHeader(status)

	if value != nil {
		_ = json.NewEncoder(writer).Encode(value)
	}
}
