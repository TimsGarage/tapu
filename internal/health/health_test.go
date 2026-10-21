package health_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/timsgarage/tapu/internal/health"
)

type fakePinger struct {
	err error
}

func (p fakePinger) PingContext(_ context.Context) error {
	return p.err
}

func TestLive(t *testing.T) {
	handler := health.NewHandler(fakePinger{})

	request := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	response := httptest.NewRecorder()

	handler.Live(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
}

func TestReadyWhenDatabaseIsAvailable(t *testing.T) {
	handler := health.NewHandler(fakePinger{})

	request := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	response := httptest.NewRecorder()

	handler.Ready(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
}

func TestReadyWhenDatabaseIsUnavailable(t *testing.T) {
	handler := health.NewHandler(fakePinger{
		err: errors.New("database unavailable"),
	})

	request := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	response := httptest.NewRecorder()

	handler.Ready(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", response.Code)
	}
}
