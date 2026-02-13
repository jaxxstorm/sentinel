package diff

import (
	"context"
	"sort"
	"time"

	"github.com/jaxxstorm/sentinel/internal/event"
	"github.com/jaxxstorm/sentinel/internal/snapshot"
)

type PeerChangeDetector struct {
	now func() time.Time
}

func NewPeerChangeDetector() *PeerChangeDetector {
	return &PeerChangeDetector{now: time.Now}
}

func (d *PeerChangeDetector) Name() string { return "peer_changes" }

func (d *PeerChangeDetector) Detect(_ context.Context, before, after snapshot.Snapshot) ([]event.Event, error) {
	prev := snapshot.IndexByPeerID(before)
	next := snapshot.IndexByPeerID(after)
	result := make([]event.Event, 0)

	nextIDs := sortedPeerIDs(next)
	for _, id := range nextIDs {
		p := next[id]
		old, exists := prev[id]
		if !exists {
			result = append(result, event.NewPeerEvent(
				event.TypePeerAdded,
				id,
				before.Hash,
				after.Hash,
				map[string]any{
					"name":   p.Name,
					"online": p.Online,
					"tags":   p.Tags,
					"routes": p.Routes,
				},
				d.now(),
			))
			continue
		}

		if !stringSliceEqual(old.Routes, p.Routes) {
			result = append(result, event.NewPeerEvent(
				event.TypePeerRoutesChanged,
				id,
				before.Hash,
				after.Hash,
				map[string]any{
					"name":          p.Name,
					"before_routes": old.Routes,
					"after_routes":  p.Routes,
				},
				d.now(),
			))
		}
		if !stringSliceEqual(old.Tags, p.Tags) {
			result = append(result, event.NewPeerEvent(
				event.TypePeerTagsChanged,
				id,
				before.Hash,
				after.Hash,
				map[string]any{
					"name":        p.Name,
					"before_tags": old.Tags,
					"after_tags":  p.Tags,
				},
				d.now(),
			))
		}
		if old.MachineAuthorized != p.MachineAuthorized {
			result = append(result, event.NewPeerEvent(
				event.TypePeerMachineAuthorizedChanged,
				id,
				before.Hash,
				after.Hash,
				map[string]any{
					"name":               p.Name,
					"before_authorized":  old.MachineAuthorized,
					"after_authorized":   p.MachineAuthorized,
					"machine_authorized": p.MachineAuthorized,
				},
				d.now(),
			))
		}
		if old.KeyExpiry != p.KeyExpiry {
			result = append(result, event.NewPeerEvent(
				event.TypePeerKeyExpiryChanged,
				id,
				before.Hash,
				after.Hash,
				map[string]any{
					"name":              p.Name,
					"before_key_expiry": old.KeyExpiry,
					"after_key_expiry":  p.KeyExpiry,
				},
				d.now(),
			))
		}
		if !old.Expired && p.Expired {
			result = append(result, event.NewPeerEvent(
				event.TypePeerKeyExpired,
				id,
				before.Hash,
				after.Hash,
				map[string]any{
					"name":       p.Name,
					"key_expiry": p.KeyExpiry,
				},
				d.now(),
			))
		}
		if old.HostinfoHash != "" && p.HostinfoHash != "" && old.HostinfoHash != p.HostinfoHash {
			result = append(result, event.NewPeerEvent(
				event.TypePeerHostinfoChanged,
				id,
				before.Hash,
				after.Hash,
				map[string]any{
					"name":                 p.Name,
					"before_hostinfo_hash": old.HostinfoHash,
					"after_hostinfo_hash":  p.HostinfoHash,
				},
				d.now(),
			))
		}
	}

	prevIDs := sortedPeerIDs(prev)
	for _, id := range prevIDs {
		old := prev[id]
		if _, exists := next[id]; exists {
			continue
		}
		result = append(result, event.NewPeerEvent(
			event.TypePeerRemoved,
			id,
			before.Hash,
			after.Hash,
			map[string]any{"name": old.Name, "reason": "missing_from_netmap"},
			d.now(),
		))
	}
	return result, nil
}

func sortedPeerIDs(peers map[string]snapshot.Peer) []string {
	ids := make([]string, 0, len(peers))
	for id := range peers {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
