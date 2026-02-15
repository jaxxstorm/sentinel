package diff

import "github.com/jaxxstorm/sentinel/internal/snapshot"

func deviceIdentityPayload(p snapshot.Peer) map[string]any {
	return map[string]any{
		"name":   p.Name,
		"tags":   normalizedIdentitySlice(p.Tags),
		"owners": normalizedIdentitySlice(p.Owners),
		"ips":    normalizedIdentitySlice(p.IPs),
	}
}

func mergePayload(base map[string]any, extras map[string]any) map[string]any {
	out := make(map[string]any, len(base)+len(extras))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extras {
		out[k] = v
	}
	return out
}

func normalizedIdentitySlice(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	out := append([]string(nil), values...)
	return out
}
