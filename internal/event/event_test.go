package event

import (
	"testing"
	"time"
)

func TestDeriveIdempotencyKeyDiffersForDifferentTimestamps(t *testing.T) {
	e1 := NewPresenceEvent(
		TypePeerOffline,
		"peer1",
		"before",
		"after",
		map[string]any{"name": "peer1"},
		time.Date(2026, 2, 12, 21, 10, 47, 0, time.UTC),
	)
	e2 := NewPresenceEvent(
		TypePeerOffline,
		"peer1",
		"before",
		"after",
		map[string]any{"name": "peer1"},
		time.Date(2026, 2, 12, 21, 11, 47, 0, time.UTC),
	)
	if DeriveIdempotencyKey(e1) == DeriveIdempotencyKey(e2) {
		t.Fatal("expected different idempotency keys for events at different timestamps")
	}
}

func TestDeriveIdempotencyKeyStableForSameEvent(t *testing.T) {
	e := NewPresenceEvent(
		TypePeerOnline,
		"peer1",
		"before",
		"after",
		map[string]any{"name": "peer1"},
		time.Date(2026, 2, 12, 21, 10, 47, 123456789, time.UTC),
	)
	k1 := DeriveIdempotencyKey(e)
	k2 := DeriveIdempotencyKey(e)
	if k1 != k2 {
		t.Fatalf("expected stable idempotency key, got %q != %q", k1, k2)
	}
}
