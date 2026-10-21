package authsdk

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrSessionNotFound = errors.New("application session not found")

type Identity struct {
	Subject          string    `json:"subject"`
	Username         string    `json:"username,omitempty"`
	ClientID         string    `json:"clientId"`
	Audience         string    `json:"audience"`
	Scope            string    `json:"scope"`
	AuthenticatedAt  time.Time `json:"authenticatedAt"`
	SessionExpiresAt time.Time `json:"sessionExpiresAt"`
}

type Session struct {
	ID          string
	Identity    Identity
	AccessToken string
	CSRFToken   string
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

type SessionStore interface {
	Put(ctx context.Context, session Session) error
	Get(ctx context.Context, id string) (*Session, error)
	Delete(ctx context.Context, id string) error
}

type MemorySessionStore struct {
	mutex    sync.RWMutex
	sessions map[string]Session
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions: make(map[string]Session),
	}
}

func (s *MemorySessionStore) Put(
	ctx context.Context,
	session Session,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.deleteExpiredLocked(time.Now().UTC())
	s.sessions[session.ID] = session

	return nil
}

func (s *MemorySessionStore) Get(
	ctx context.Context,
	id string,
) (*Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	s.mutex.RLock()
	found, exists := s.sessions[id]
	s.mutex.RUnlock()

	if !exists {
		return nil, ErrSessionNotFound
	}

	if !found.ExpiresAt.After(now) {
		s.mutex.Lock()
		delete(s.sessions, id)
		s.mutex.Unlock()

		return nil, ErrSessionNotFound
	}

	copy := found
	return &copy, nil
}

func (s *MemorySessionStore) Delete(
	ctx context.Context,
	id string,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.sessions, id)
	return nil
}

func (s *MemorySessionStore) deleteExpiredLocked(
	now time.Time,
) {
	for id, session := range s.sessions {
		if !session.ExpiresAt.After(now) {
			delete(s.sessions, id)
		}
	}
}
