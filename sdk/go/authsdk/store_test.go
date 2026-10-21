package authsdk_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/timsgarage/tapu/sdk/go/authsdk"
)

func TestMemorySessionStore(t *testing.T) {
	store := authsdk.NewMemorySessionStore()

	session := authsdk.Session{
		ID:        "test-session",
		ExpiresAt: time.Now().UTC().Add(time.Minute),
	}

	if err := store.Put(
		context.Background(),
		session,
	); err != nil {
		t.Fatalf("Put() returned an error: %v", err)
	}

	found, err := store.Get(
		context.Background(),
		session.ID,
	)
	if err != nil {
		t.Fatalf("Get() returned an error: %v", err)
	}

	if found.ID != session.ID {
		t.Fatalf(
			"expected session %q, got %q",
			session.ID,
			found.ID,
		)
	}

	if err := store.Delete(
		context.Background(),
		session.ID,
	); err != nil {
		t.Fatalf("Delete() returned an error: %v", err)
	}

	_, err = store.Get(
		context.Background(),
		session.ID,
	)

	if !errors.Is(err, authsdk.ErrSessionNotFound) {
		t.Fatalf(
			"expected ErrSessionNotFound, got %v",
			err,
		)
	}
}

func TestMemorySessionStoreRejectsExpiredSession(
	t *testing.T,
) {
	store := authsdk.NewMemorySessionStore()

	session := authsdk.Session{
		ID:        "expired-session",
		ExpiresAt: time.Now().UTC().Add(-time.Minute),
	}

	if err := store.Put(
		context.Background(),
		session,
	); err != nil {
		t.Fatalf("Put() returned an error: %v", err)
	}

	_, err := store.Get(
		context.Background(),
		session.ID,
	)

	if !errors.Is(err, authsdk.ErrSessionNotFound) {
		t.Fatalf(
			"expected ErrSessionNotFound, got %v",
			err,
		)
	}
}
