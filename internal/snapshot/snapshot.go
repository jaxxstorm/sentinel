package snapshot

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"time"

	"github.com/jaxxstorm/sentinel/internal/source"
)

type Peer struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Online            bool              `json:"online"`
	Tags              []string          `json:"tags,omitempty"`
	Routes            []string          `json:"routes,omitempty"`
	MachineAuthorized bool              `json:"machine_authorized,omitempty"`
	Expired           bool              `json:"expired,omitempty"`
	KeyExpiry         string            `json:"key_expiry,omitempty"`
	HostinfoHash      string            `json:"hostinfo_hash,omitempty"`
	Meta              map[string]string `json:"meta,omitempty"`
}

type Prefs struct {
	AdvertiseRoutes []string `json:"advertise_routes,omitempty"`
	ExitNodeID      string   `json:"exit_node_id,omitempty"`
	RunSSH          bool     `json:"run_ssh,omitempty"`
	ShieldsUp       bool     `json:"shields_up,omitempty"`
}

type Tailnet struct {
	Domain     string `json:"domain,omitempty"`
	TKAEnabled bool   `json:"tka_enabled,omitempty"`
}

type Snapshot struct {
	CapturedAt    time.Time `json:"captured_at"`
	Peers         []Peer    `json:"peers"`
	DaemonState   string    `json:"daemon_state,omitempty"`
	Prefs         Prefs     `json:"prefs,omitempty"`
	Tailnet       Tailnet   `json:"tailnet,omitempty"`
	LastErrorText string    `json:"last_error_text,omitempty"`
	Hash          string    `json:"hash"`
}

func Normalize(nm source.Netmap, now time.Time) Snapshot {
	peers := make([]Peer, 0, len(nm.Peers))
	for _, p := range nm.Peers {
		tags := append([]string(nil), p.Tags...)
		sort.Strings(tags)
		routes := append([]string(nil), p.Routes...)
		sort.Strings(routes)
		peers = append(peers, Peer{
			ID:                p.ID,
			Name:              p.Name,
			Online:            p.Online,
			Tags:              tags,
			Routes:            routes,
			MachineAuthorized: p.MachineAuthorized,
			Expired:           p.Expired,
			KeyExpiry:         p.KeyExpiry,
			HostinfoHash:      p.HostinfoHash,
			Meta:              redactVolatileMeta(p.Metadata),
		})
	}
	sort.Slice(peers, func(i, j int) bool { return peers[i].ID < peers[j].ID })

	advertiseRoutes := append([]string(nil), nm.Prefs.AdvertiseRoutes...)
	sort.Strings(advertiseRoutes)
	s := Snapshot{
		CapturedAt:  now.UTC(),
		Peers:       peers,
		DaemonState: nm.DaemonState,
		Prefs: Prefs{
			AdvertiseRoutes: advertiseRoutes,
			ExitNodeID:      nm.Prefs.ExitNodeID,
			RunSSH:          nm.Prefs.RunSSH,
			ShieldsUp:       nm.Prefs.ShieldsUp,
		},
		Tailnet: Tailnet{
			Domain:     nm.Tailnet.Domain,
			TKAEnabled: nm.Tailnet.TKAEnabled,
		},
		LastErrorText: nm.ErrorMessage,
	}
	s.Hash = Hash(s)
	return s
}

func Hash(s Snapshot) string {
	normalized := struct {
		Peers         []Peer  `json:"peers"`
		DaemonState   string  `json:"daemon_state,omitempty"`
		Prefs         Prefs   `json:"prefs,omitempty"`
		Tailnet       Tailnet `json:"tailnet,omitempty"`
		LastErrorText string  `json:"last_error_text,omitempty"`
	}{
		Peers:         s.Peers,
		DaemonState:   s.DaemonState,
		Prefs:         s.Prefs,
		Tailnet:       s.Tailnet,
		LastErrorText: s.LastErrorText,
	}
	b, _ := json.Marshal(normalized)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func IndexByPeerID(s Snapshot) map[string]Peer {
	m := make(map[string]Peer, len(s.Peers))
	for _, p := range s.Peers {
		m[p.ID] = p
	}
	return m
}

func redactVolatileMeta(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		// Skip volatile network path details to avoid spurious diffs in v0.
		if k == "endpoint" || k == "derp" || k == "relay_path" {
			continue
		}
		out[k] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
