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
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Online bool              `json:"online"`
	Tags   []string          `json:"tags,omitempty"`
	Meta   map[string]string `json:"meta,omitempty"`
}

type Snapshot struct {
	CapturedAt time.Time `json:"captured_at"`
	Peers      []Peer    `json:"peers"`
	Hash       string    `json:"hash"`
}

func Normalize(nm source.Netmap, now time.Time) Snapshot {
	peers := make([]Peer, 0, len(nm.Peers))
	for _, p := range nm.Peers {
		tags := append([]string(nil), p.Tags...)
		sort.Strings(tags)
		peers = append(peers, Peer{
			ID:     p.ID,
			Name:   p.Name,
			Online: p.Online,
			Tags:   tags,
			Meta:   redactVolatileMeta(p.Metadata),
		})
	}
	sort.Slice(peers, func(i, j int) bool { return peers[i].ID < peers[j].ID })

	s := Snapshot{CapturedAt: now.UTC(), Peers: peers}
	s.Hash = Hash(s)
	return s
}

func Hash(s Snapshot) string {
	normalized := struct {
		Peers []Peer `json:"peers"`
	}{Peers: s.Peers}
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
