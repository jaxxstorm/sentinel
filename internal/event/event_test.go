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

func TestNewEventSubjectTypeHelpers(t *testing.T) {
	now := time.Date(2026, 2, 13, 10, 0, 0, 0, time.UTC)
	peer := NewPeerEvent(TypePeerAdded, "peer-a", "before", "after", nil, now)
	daemon := NewDaemonEvent(TypeDaemonStateChanged, "local", "before", "after", nil, now)
	prefs := NewPrefsEvent(TypePrefsRunSSHChanged, "local", "before", "after", nil, now)
	tailnet := NewTailnetEvent(TypeTailnetDomainChanged, "tailnet", "before", "after", nil, now)

	if peer.SubjectType != SubjectPeer {
		t.Fatalf("expected peer subject type %q, got %q", SubjectPeer, peer.SubjectType)
	}
	if daemon.SubjectType != SubjectDaemon {
		t.Fatalf("expected daemon subject type %q, got %q", SubjectDaemon, daemon.SubjectType)
	}
	if prefs.SubjectType != SubjectPrefs {
		t.Fatalf("expected prefs subject type %q, got %q", SubjectPrefs, prefs.SubjectType)
	}
	if tailnet.SubjectType != SubjectTailnet {
		t.Fatalf("expected tailnet subject type %q, got %q", SubjectTailnet, tailnet.SubjectType)
	}
}

func TestIsKnownType(t *testing.T) {
	known := []string{
		TypePeerOnline,
		TypePeerOffline,
		TypePeerAdded,
		TypePeerRemoved,
		TypePeerRoutesChanged,
		TypePeerTagsChanged,
		TypePeerMachineAuthorizedChanged,
		TypePeerKeyExpiryChanged,
		TypePeerKeyExpired,
		TypePeerHostinfoChanged,
		TypeDaemonStateChanged,
		TypePrefsAdvertiseRoutesChanged,
		TypePrefsExitNodeChanged,
		TypePrefsRunSSHChanged,
		TypePrefsShieldsUpChanged,
		TypeTailnetDomainChanged,
		TypeTailnetTKAEnabledChanged,
	}
	for _, et := range known {
		if !IsKnownType(et) {
			t.Fatalf("expected event type %q to be known", et)
		}
	}
	if IsKnownType("peer.custom.unsupported") {
		t.Fatal("expected unknown type to be rejected")
	}
}
