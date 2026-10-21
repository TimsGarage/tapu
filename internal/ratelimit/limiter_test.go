package ratelimit_test

import (
	"testing"
	"time"

	"github.com/timsgarage/tapu/internal/ratelimit"
)

func TestLimiterRejectsAfterMaximumAttempts(t *testing.T) {
	limiter := ratelimit.New(2, time.Minute)

	if !limiter.Allow("username:alice", "ip:127.0.0.1") {
		t.Fatal("expected first request to be allowed")
	}

	if !limiter.Allow("username:alice", "ip:127.0.0.1") {
		t.Fatal("expected second request to be allowed")
	}

	if limiter.Allow("username:alice", "ip:127.0.0.1") {
		t.Fatal("expected third request to be rejected")
	}
}

func TestLimiterReset(t *testing.T) {
	limiter := ratelimit.New(1, time.Minute)

	if !limiter.Allow("username:alice") {
		t.Fatal("expected first request to be allowed")
	}

	if limiter.Allow("username:alice") {
		t.Fatal("expected second request to be rejected")
	}

	limiter.Reset("username:alice")

	if !limiter.Allow("username:alice") {
		t.Fatal("expected request after reset to be allowed")
	}
}
