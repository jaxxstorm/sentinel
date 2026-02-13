package snapshot

import (
	"testing"
	"time"

	"github.com/jaxxstorm/sentinel/internal/source"
)

func TestNormalizeIgnoresVolatileMetaFieldsForHash(t *testing.T) {
	now := time.Date(2026, 2, 13, 0, 0, 0, 0, time.UTC)
	a := Normalize(source.Netmap{
		Peers: []source.Peer{{
			ID:     "peer-1",
			Name:   "peer-1",
			Online: true,
			Metadata: map[string]string{
				"os":         "linux",
				"endpoint":   "1.2.3.4:1234",
				"relay_path": "derp-1",
			},
		}},
	}, now)
	b := Normalize(source.Netmap{
		Peers: []source.Peer{{
			ID:     "peer-1",
			Name:   "peer-1",
			Online: true,
			Metadata: map[string]string{
				"os":         "linux",
				"endpoint":   "9.9.9.9:4321",
				"relay_path": "derp-2",
			},
		}},
	}, now)

	if a.Hash != b.Hash {
		t.Fatalf("expected volatile metadata changes to preserve hash, got %q != %q", a.Hash, b.Hash)
	}
}

func TestNormalizeHashChangesForRouteUpdates(t *testing.T) {
	now := time.Date(2026, 2, 13, 0, 0, 0, 0, time.UTC)
	a := Normalize(source.Netmap{
		Peers: []source.Peer{{
			ID:     "peer-1",
			Name:   "peer-1",
			Online: true,
			Routes: []string{"10.0.0.0/24"},
		}},
	}, now)
	b := Normalize(source.Netmap{
		Peers: []source.Peer{{
			ID:     "peer-1",
			Name:   "peer-1",
			Online: true,
			Routes: []string{"10.1.0.0/24"},
		}},
	}, now)

	if a.Hash == b.Hash {
		t.Fatalf("expected route changes to alter hash, got %q", a.Hash)
	}
}

func TestNormalizeHashChangesForRuntimeFields(t *testing.T) {
	now := time.Date(2026, 2, 13, 0, 0, 0, 0, time.UTC)
	a := Normalize(source.Netmap{
		DaemonState: "Starting",
		Prefs: source.Prefs{
			RunSSH:          false,
			ShieldsUp:       false,
			ExitNodeID:      "",
			AdvertiseRoutes: []string{"10.0.0.0/24"},
		},
		Tailnet: source.Tailnet{
			Domain:     "tail-a",
			TKAEnabled: false,
		},
	}, now)
	b := Normalize(source.Netmap{
		DaemonState: "Running",
		Prefs: source.Prefs{
			RunSSH:          true,
			ShieldsUp:       true,
			ExitNodeID:      "node-a",
			AdvertiseRoutes: []string{"10.1.0.0/24"},
		},
		Tailnet: source.Tailnet{
			Domain:     "tail-b",
			TKAEnabled: true,
		},
	}, now)

	if a.Hash == b.Hash {
		t.Fatalf("expected runtime field changes to alter hash, got %q", a.Hash)
	}
}
