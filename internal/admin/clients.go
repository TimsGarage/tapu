package admin

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/timsgarage/tapu/internal/clientsecret"
	"gorm.io/gorm"
)

type clientResponse struct {
	ID           uuid.UUID   `json:"id"`
	ClientID     string      `json:"clientId"`
	Name         string      `json:"name"`
	ClientType   string      `json:"clientType"`
	Enabled      bool        `json:"enabled"`
	RedirectURIs []string    `json:"redirectUris"`
	ServiceIDs   []uuid.UUID `json:"serviceIds"`
	CreatedAt    time.Time   `json:"createdAt"`
}

type createdClientResponse struct {
	clientResponse
	ClientSecret string `json:"clientSecret,omitempty"`
}

func (h *Handler) ListClients(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var clients []OAuthClient

	if err := h.database.WithContext(request.Context()).
		Order("name ASC").
		Find(&clients).
		Error; err != nil {
		h.internalError(writer, request, err)
		return
	}

	response := make([]clientResponse, 0, len(clients))

	for _, client := range clients {
		item, err := h.clientResponse(request, client)
		if err != nil {
			h.internalError(writer, request, err)
			return
		}

		response = append(response, item)
	}

	writeJSON(writer, http.StatusOK, response)
}

func (h *Handler) CreateClient(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var input struct {
		Name         string      `json:"name"`
		ClientType   string      `json:"clientType"`
		RedirectURIs []string    `json:"redirectUris"`
		ServiceIDs   []uuid.UUID `json:"serviceIds"`
	}

	if err := decodeJSON(writer, request, &input); err != nil {
		writeAPIError(writer, http.StatusBadRequest, "Invalid request")
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.ClientType = strings.TrimSpace(input.ClientType)

	if err := validateName(input.Name); err != nil {
		writeAPIError(writer, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateClientType(input.ClientType); err != nil {
		writeAPIError(writer, http.StatusBadRequest, err.Error())
		return
	}

	redirectURIs := uniqueStrings(input.RedirectURIs)
	for _, redirectURI := range redirectURIs {
		if err := validateRedirectURI(redirectURI); err != nil {
			writeAPIError(writer, http.StatusBadRequest, err.Error())
			return
		}
	}

	if len(redirectURIs) == 0 {
		writeAPIError(
			writer,
			http.StatusBadRequest,
			"At least one redirect URI is required",
		)
		return
	}

	clientID, _, err := clientsecret.Generate()
	if err != nil {
		h.internalError(writer, request, err)
		return
	}

	var rawSecret string
	var secretHash []byte

	if input.ClientType == ClientTypeConfidential {
		rawSecret, secretHash, err = clientsecret.Generate()
		if err != nil {
			h.internalError(writer, request, err)
			return
		}
	}

	now := time.Now().UTC()
	client := OAuthClient{
		ID:         uuid.New(),
		ClientID:   clientID,
		Name:       input.Name,
		ClientType: input.ClientType,
		SecretHash: secretHash,
		Enabled:    true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	err = h.database.WithContext(request.Context()).
		Transaction(func(transaction *gorm.DB) error {
			if err := transaction.Create(&client).Error; err != nil {
				return err
			}

			for _, redirectURI := range redirectURIs {
				model := OAuthClientRedirectURI{
					ID:            uuid.New(),
					OAuthClientID: client.ID,
					RedirectURI:   redirectURI,
				}

				if err := transaction.Create(&model).Error; err != nil {
					return err
				}
			}

			for _, serviceID := range uniqueUUIDs(input.ServiceIDs) {
				model := OAuthClientService{
					OAuthClientID: client.ID,
					ServiceID:     serviceID,
					CreatedAt:     now,
				}

				if err := transaction.Create(&model).Error; err != nil {
					return err
				}
			}

			return nil
		})

	if err != nil {
		if isForeignKeyViolation(err) {
			writeAPIError(
				writer,
				http.StatusBadRequest,
				"One or more services do not exist",
			)
			return
		}

		h.internalError(writer, request, err)
		return
	}

	item, err := h.clientResponse(request, client)
	if err != nil {
		h.internalError(writer, request, err)
		return
	}

	h.recordAudit(
		request,
		"oauth_client.created",
		"oauth_client",
		&client.ID,
		map[string]any{
			"clientId":   client.ClientID,
			"clientType": client.ClientType,
		},
	)

	writeJSON(writer, http.StatusCreated, createdClientResponse{
		clientResponse: item,
		ClientSecret:   rawSecret,
	})
}

func (h *Handler) UpdateClient(
	writer http.ResponseWriter,
	request *http.Request,
) {
	clientID, err := parseUUID(chi.URLParam(request, "clientID"))
	if err != nil {
		writeAPIError(writer, http.StatusBadRequest, "Invalid client ID")
		return
	}

	var input struct {
		Name         *string      `json:"name"`
		Enabled      *bool        `json:"enabled"`
		RedirectURIs *[]string    `json:"redirectUris"`
		ServiceIDs   *[]uuid.UUID `json:"serviceIds"`
	}

	if err := decodeJSON(writer, request, &input); err != nil {
		writeAPIError(writer, http.StatusBadRequest, "Invalid request")
		return
	}

	err = h.database.WithContext(request.Context()).
		Transaction(func(transaction *gorm.DB) error {
			updates := map[string]any{
				"updated_at": time.Now().UTC(),
			}

			if input.Name != nil {
				name := strings.TrimSpace(*input.Name)
				if err := validateName(name); err != nil {
					return err
				}
				updates["name"] = name
			}

			if input.Enabled != nil {
				updates["enabled"] = *input.Enabled
			}

			result := transaction.
				Model(&OAuthClient{}).
				Where("id = ?", clientID).
				Updates(updates)

			if result.Error != nil {
				return result.Error
			}

			if result.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}

			if input.RedirectURIs != nil {
				values := uniqueStrings(*input.RedirectURIs)
				if len(values) == 0 {
					return errRedirectURIsRequired
				}

				for _, value := range values {
					if err := validateRedirectURI(value); err != nil {
						return err
					}
				}

				if err := transaction.
					Delete(
						&OAuthClientRedirectURI{},
						"oauth_client_id = ?",
						clientID,
					).
					Error; err != nil {
					return err
				}

				for _, value := range values {
					redirect := OAuthClientRedirectURI{
						ID:            uuid.New(),
						OAuthClientID: clientID,
						RedirectURI:   value,
					}

					if err := transaction.Create(&redirect).Error; err != nil {
						return err
					}
				}
			}

			if input.ServiceIDs != nil {
				if err := transaction.
					Delete(
						&OAuthClientService{},
						"oauth_client_id = ?",
						clientID,
					).
					Error; err != nil {
					return err
				}

				for _, serviceID := range uniqueUUIDs(*input.ServiceIDs) {
					association := OAuthClientService{
						OAuthClientID: clientID,
						ServiceID:     serviceID,
						CreatedAt:     time.Now().UTC(),
					}

					if err := transaction.Create(&association).Error; err != nil {
						return err
					}
				}
			}

			return nil
		})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			writeAPIError(writer, http.StatusNotFound, "Client not found")
			return
		}

		if err == errRedirectURIsRequired {
			writeAPIError(
				writer,
				http.StatusBadRequest,
				err.Error(),
			)
			return
		}

		if isValidationError(err) {
			writeAPIError(writer, http.StatusBadRequest, err.Error())
			return
		}

		h.internalError(writer, request, err)
		return
	}

	var client OAuthClient
	if err := h.database.WithContext(request.Context()).
		First(&client, "id = ?", clientID).
		Error; err != nil {
		h.internalError(writer, request, err)
		return
	}

	item, err := h.clientResponse(request, client)
	if err != nil {
		h.internalError(writer, request, err)
		return
	}

	h.recordAudit(
		request,
		"oauth_client.updated",
		"oauth_client",
		&clientID,
		nil,
	)

	writeJSON(writer, http.StatusOK, item)
}

func (h *Handler) RotateClientSecret(
	writer http.ResponseWriter,
	request *http.Request,
) {
	clientID, err := parseUUID(chi.URLParam(request, "clientID"))
	if err != nil {
		writeAPIError(writer, http.StatusBadRequest, "Invalid client ID")
		return
	}

	rawSecret, secretHash, err := clientsecret.Generate()
	if err != nil {
		h.internalError(writer, request, err)
		return
	}

	result := h.database.WithContext(request.Context()).
		Model(&OAuthClient{}).
		Where(
			"id = ? AND client_type = ?",
			clientID,
			ClientTypeConfidential,
		).
		Updates(map[string]any{
			"secret_hash": secretHash,
			"updated_at":  time.Now().UTC(),
		})

	if result.Error != nil {
		h.internalError(writer, request, result.Error)
		return
	}

	if result.RowsAffected == 0 {
		writeAPIError(
			writer,
			http.StatusNotFound,
			"Confidential client not found",
		)
		return
	}

	h.recordAudit(
		request,
		"oauth_client.secret_rotated",
		"oauth_client",
		&clientID,
		nil,
	)

	writeJSON(writer, http.StatusOK, map[string]string{
		"clientSecret": rawSecret,
	})
}

func (h *Handler) clientResponse(
	request *http.Request,
	client OAuthClient,
) (clientResponse, error) {
	var redirectURIs []string
	var serviceIDs []uuid.UUID

	if err := h.database.WithContext(request.Context()).
		Model(&OAuthClientRedirectURI{}).
		Where("oauth_client_id = ?", client.ID).
		Order("redirect_uri ASC").
		Pluck("redirect_uri", &redirectURIs).
		Error; err != nil {
		return clientResponse{}, err
	}

	if err := h.database.WithContext(request.Context()).
		Model(&OAuthClientService{}).
		Where("oauth_client_id = ?", client.ID).
		Pluck("service_id", &serviceIDs).
		Error; err != nil {
		return clientResponse{}, err
	}

	return clientResponse{
		ID:           client.ID,
		ClientID:     client.ClientID,
		Name:         client.Name,
		ClientType:   client.ClientType,
		Enabled:      client.Enabled,
		RedirectURIs: redirectURIs,
		ServiceIDs:   serviceIDs,
		CreatedAt:    client.CreatedAt,
	}, nil
}
