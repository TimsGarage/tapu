package ratelimit

import (
	"sync"
	"time"
)

type entry struct {
	windowStarted time.Time
	attempts      int
	lastSeen      time.Time
}

type Limiter struct {
	mutex         sync.Mutex
	maxAttempts   int
	window        time.Duration
	entries       map[string]entry
	lastCleanupAt time.Time
}

func New(maxAttempts int, window time.Duration) *Limiter {
	return &Limiter{
		maxAttempts:   maxAttempts,
		window:        window,
		entries:       make(map[string]entry),
		lastCleanupAt: time.Now(),
	}
}

func (l *Limiter) Allow(keys ...string) bool {
	now := time.Now()

	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.cleanup(now)

	for _, key := range keys {
		current := l.entries[key]

		if current.windowStarted.IsZero() ||
			now.Sub(current.windowStarted) >= l.window {
			continue
		}

		if current.attempts >= l.maxAttempts {
			return false
		}
	}

	for _, key := range keys {
		current := l.entries[key]

		if current.windowStarted.IsZero() ||
			now.Sub(current.windowStarted) >= l.window {
			current = entry{
				windowStarted: now,
			}
		}

		current.attempts++
		current.lastSeen = now
		l.entries[key] = current
	}

	return true
}

func (l *Limiter) Reset(keys ...string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for _, key := range keys {
		delete(l.entries, key)
	}
}

func (l *Limiter) cleanup(now time.Time) {
	if now.Sub(l.lastCleanupAt) < l.window {
		return
	}

	for key, current := range l.entries {
		if now.Sub(current.lastSeen) >= 2*l.window {
			delete(l.entries, key)
		}
	}

	l.lastCleanupAt = now
}
