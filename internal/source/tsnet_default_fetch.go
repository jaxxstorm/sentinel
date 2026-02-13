package source

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"tailscale.com/client/local"
	"tailscale.com/ipn"
	"tailscale.com/tsnet"
)

// DefaultTSNetFetch converts tsnet local status JSON into Sentinel's netmap shape.
func DefaultTSNetFetch(ctx context.Context, server *tsnet.Server) (Netmap, error) {
	client, err := server.LocalClient()
	if err != nil {
		return Netmap{}, fmt.Errorf("create local client: %w", err)
	}
	status, err := client.Status(ctx)
	if err != nil {
		return Netmap{}, fmt.Errorf("fetch tailscale status: %w", err)
	}

	statusData, err := json.Marshal(status)
	if err != nil {
		return Netmap{}, fmt.Errorf("marshal status: %w", err)
	}
	nm, err := decodeNetmapFromStatusJSON(statusData)
	if err != nil {
		return Netmap{}, fmt.Errorf("decode status: %w", err)
	}

	// Fallback to initial NetMap when /status does not expose peers.
	if len(nm.Peers) == 0 {
		netmap, err := fetchNetmapFromIPNBus(ctx, client)
		if err == nil && len(netmap.Peers) > 0 {
			nm = netmap
		}
	}

	nm.PolledAt = time.Now().UTC()
	return nm, nil
}

func fetchNetmapFromIPNBus(ctx context.Context, client *local.Client) (Netmap, error) {
	watchCtx := ctx
	cancel := func() {}
	if _, ok := ctx.Deadline(); !ok {
		watchCtx, cancel = context.WithTimeout(ctx, 3*time.Second)
	}
	defer cancel()

	watcher, err := client.WatchIPNBus(watchCtx, ipn.NotifyInitialState|ipn.NotifyInitialNetMap)
	if err != nil {
		return Netmap{}, err
	}
	defer watcher.Close()

	for {
		note, err := watcher.Next()
		if err != nil {
			return Netmap{}, err
		}
		if note.State != nil {
			// Retain state if netmap arrives in a later frame.
			// LocalClient initial frames can be split by field.
		}
		if note.NetMap == nil {
			continue
		}
		netmapData, err := json.Marshal(note.NetMap)
		if err != nil {
			return Netmap{}, err
		}
		nm, err := decodeNetMapJSON(netmapData)
		if err != nil {
			return Netmap{}, err
		}
		if note.State != nil {
			nm.DaemonState = note.State.String()
		}
		return nm, nil
	}
}

func decodePeersFromStatusJSON(data []byte) ([]Peer, error) {
	nm, err := decodeNetmapFromStatusJSON(data)
	if err != nil {
		return nil, err
	}
	return nm.Peers, nil
}

func decodeNetmapFromStatusJSON(data []byte) (Netmap, error) {
	var raw struct {
		BackendState   string                    `json:"BackendState"`
		CurrentTailnet map[string]any            `json:"CurrentTailnet"`
		Peer           map[string]map[string]any `json:"Peer"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return Netmap{}, err
	}

	nm := Netmap{DaemonState: strings.TrimSpace(raw.BackendState)}
	if raw.CurrentTailnet != nil {
		nm.Tailnet.Domain = firstNonEmpty(
			stringVal(raw.CurrentTailnet, "Name"),
			stringVal(raw.CurrentTailnet, "MagicDNSSuffix"),
		)
	}

	peers := make([]Peer, 0, len(raw.Peer))
	for fallbackID, peer := range raw.Peer {
		p := Peer{
			ID:                firstNonEmpty(stringVal(peer, "StableID"), fallbackID),
			Name:              firstNonEmpty(stringVal(peer, "HostName"), hostFromDNSName(stringVal(peer, "DNSName"))),
			Online:            boolVal(peer, "Online"),
			Tags:              sortedCopy(stringSliceVal(peer, "Tags")),
			Routes:            sortedCopy(stringSliceVal(peer, "PrimaryRoutes")),
			MachineAuthorized: boolVal(peer, "MachineAuthorized"),
			Expired:           boolVal(peer, "Expired"),
			KeyExpiry:         anyToString(peer["KeyExpiry"]),
		}
		meta := map[string]string{}
		if v := stringVal(peer, "OS"); v != "" {
			meta["os"] = v
		}
		if hostinfo := mapVal(peer, "Hostinfo"); hostinfo != nil {
			p.HostinfoHash = stableMapHash(hostinfo)
		}
		if v := anyToString(peer["UserID"]); v != "" {
			meta["user_id"] = v
		}
		if v := anyToString(peer["LastSeen"]); v != "" {
			meta["last_seen"] = v
		}
		if len(meta) > 0 {
			p.Metadata = meta
		}
		peers = append(peers, p)
	}
	nm.Peers = peers
	return nm, nil
}

func decodePeersFromNetMapJSON(data []byte) ([]Peer, error) {
	nm, err := decodeNetMapJSON(data)
	if err != nil {
		return nil, err
	}
	return nm.Peers, nil
}

func decodeNetMapJSON(data []byte) (Netmap, error) {
	var raw struct {
		Domain     string           `json:"Domain"`
		TKAEnabled bool             `json:"TKAEnabled"`
		Peers      []map[string]any `json:"Peers"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return Netmap{}, err
	}

	nm := Netmap{
		Tailnet: Tailnet{
			Domain:     strings.TrimSpace(raw.Domain),
			TKAEnabled: raw.TKAEnabled,
		},
	}

	peers := make([]Peer, 0, len(raw.Peers))
	for _, node := range raw.Peers {
		p := Peer{
			ID:                firstNonEmpty(stringVal(node, "StableID"), anyToString(node["ID"])),
			Name:              firstNonEmpty(stringVal(node, "ComputedName"), hostFromDNSName(stringVal(node, "Name"))),
			Online:            boolVal(node, "Online"),
			Tags:              sortedCopy(stringSliceVal(node, "Tags")),
			Routes:            sortedCopy(stringSliceVal(node, "PrimaryRoutes")),
			MachineAuthorized: boolVal(node, "MachineAuthorized"),
			Expired:           boolVal(node, "Expired"),
			KeyExpiry:         anyToString(node["KeyExpiry"]),
		}
		meta := map[string]string{}
		if hostinfo := mapVal(node, "Hostinfo"); hostinfo != nil {
			p.HostinfoHash = stableMapHash(hostinfo)
			if v := stringVal(hostinfo, "OS"); v != "" {
				meta["os"] = v
			}
			if p.Name == "" {
				p.Name = stringVal(hostinfo, "Hostname")
			}
		}
		if v := anyToString(node["User"]); v != "" {
			meta["user_id"] = v
		}
		if v := anyToString(node["LastSeen"]); v != "" {
			meta["last_seen"] = v
		}
		if len(meta) > 0 {
			p.Metadata = meta
		}
		peers = append(peers, p)
	}
	nm.Peers = peers
	return nm, nil
}

func stringVal(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func boolVal(m map[string]any, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func stringSliceVal(m map[string]any, key string) []string {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func mapVal(m map[string]any, key string) map[string]any {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	out, _ := v.(map[string]any)
	return out
}

func anyToString(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case float64:
		return strconv.FormatInt(int64(x), 10)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case json.Number:
		return x.String()
	case bool:
		if x {
			return "true"
		}
		return "false"
	default:
		if s, ok := x.(fmt.Stringer); ok {
			return s.String()
		}
		return fmt.Sprint(x)
	}
}

func hostFromDNSName(s string) string {
	s = strings.TrimSpace(strings.TrimSuffix(s, "."))
	if s == "" {
		return ""
	}
	if idx := strings.IndexByte(s, '.'); idx > 0 {
		return s[:idx]
	}
	return s
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func sortedCopy(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := append([]string(nil), items...)
	sort.Strings(out)
	return out
}

func stableMapHash(m map[string]any) string {
	if len(m) == 0 {
		return ""
	}
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:8])
}
