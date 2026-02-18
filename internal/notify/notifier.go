package notify

import (
	"context"
	"errors"
	"net/netip"
	"path"
	"strings"
	"time"

	"github.com/jaxxstorm/sentinel/internal/event"
	"github.com/jaxxstorm/sentinel/internal/state"
)

type Route struct {
	EventTypes []string
	Severities []string
	Sinks      []string
	Device     DeviceSelector
	Filters    RouteFilters
}

type DeviceSelector struct {
	Names  []string
	Tags   []string
	Owners []string
	IPs    []string
}

type RouteFilters struct {
	Include NotificationFilter
	Exclude NotificationFilter
}

type NotificationFilter struct {
	DeviceNames []string
	Tags        []string
	IPs         []string
	Events      []string
}

type deviceIdentity struct {
	Name   string
	Tags   []string
	Owners []string
	IPs    []string
}

type Config struct {
	Routes            []Route
	IdempotencyKeyTTL time.Duration
}

type Notification struct {
	Event          event.Event `json:"event"`
	IdempotencyKey string      `json:"idempotency_key"`
}

type Sink interface {
	Name() string
	Send(ctx context.Context, n Notification) error
}

type Result struct {
	Sent       int
	Suppressed int
	DryRun     int
}

type Notifier struct {
	cfg   Config
	store state.StateStore
	sinks map[string]Sink
}

func New(cfg Config, store state.StateStore, sinks []Sink) *Notifier {
	m := make(map[string]Sink, len(sinks))
	for _, sink := range sinks {
		m[sink.Name()] = sink
	}
	if cfg.IdempotencyKeyTTL <= 0 {
		cfg.IdempotencyKeyTTL = 24 * time.Hour
	}
	return &Notifier{cfg: cfg, store: store, sinks: m}
}

func (n *Notifier) Notify(ctx context.Context, events []event.Event, dryRun bool) (Result, error) {
	result := Result{}
	for _, evt := range events {
		routeTargets := n.targetsFor(evt)
		if len(routeTargets) == 0 {
			continue
		}
		key := event.DeriveIdempotencyKey(evt)
		seen, err := n.store.SeenIdempotencyKey(key)
		if err != nil {
			return result, err
		}
		if seen {
			result.Suppressed++
			continue
		}

		note := Notification{Event: evt, IdempotencyKey: key}
		if dryRun {
			result.DryRun += len(routeTargets)
			if err := n.store.RecordIdempotencyKey(key, n.cfg.IdempotencyKeyTTL); err != nil {
				return result, err
			}
			continue
		}

		for _, target := range routeTargets {
			sink, ok := n.sinks[target]
			if !ok {
				continue
			}
			if err := sink.Send(ctx, note); err != nil {
				return result, err
			}
			result.Sent++
		}
		if err := n.store.RecordIdempotencyKey(key, n.cfg.IdempotencyKeyTTL); err != nil {
			return result, err
		}
	}
	return result, nil
}

func (n *Notifier) targetsFor(evt event.Event) []string {
	out := []string{}
	for _, r := range n.cfg.Routes {
		if len(r.EventTypes) > 0 && !matchesEventType(r.EventTypes, evt.EventType) {
			continue
		}
		if len(r.Severities) > 0 && !contains(r.Severities, evt.Severity) {
			continue
		}
		if !matchesRouteFilters(r.Filters, evt) {
			continue
		}
		if !matchesDeviceSelector(r.Device, evt) {
			continue
		}
		out = append(out, r.Sinks...)
	}
	return uniq(out)
}

func matchesEventType(items []string, target string) bool {
	for _, item := range items {
		if item == "*" {
			return true
		}
		if item == target {
			return true
		}
	}
	return false
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func matchesDeviceSelector(selector DeviceSelector, evt event.Event) bool {
	if !hasSelectorFilters(selector) {
		return true
	}
	identity, ok := deviceIdentityFromEvent(evt)
	if !ok {
		return false
	}
	if len(selector.Names) > 0 && !matchesSelectorString(selector.Names, identity.Name) {
		return false
	}
	if len(selector.Tags) > 0 && !matchesSelectorAny(selector.Tags, identity.Tags) {
		return false
	}
	if len(selector.Owners) > 0 && !matchesSelectorAny(selector.Owners, identity.Owners) {
		return false
	}
	if len(selector.IPs) > 0 && !matchesSelectorAnyIPs(selector.IPs, identity.IPs) {
		return false
	}
	return true
}

func hasSelectorFilters(selector DeviceSelector) bool {
	return len(selector.Names) > 0 || len(selector.Tags) > 0 || len(selector.Owners) > 0 || len(selector.IPs) > 0
}

func matchesRouteFilters(filters RouteFilters, evt event.Event) bool {
	if !hasNotificationFilters(filters.Include) && !hasNotificationFilters(filters.Exclude) {
		return true
	}
	var (
		identity   deviceIdentity
		identityOK bool
	)
	if hasIdentityFilters(filters.Include) || hasIdentityFilters(filters.Exclude) {
		identity, identityOK = deviceIdentityFromEvent(evt)
		if !identityOK {
			return false
		}
	}
	if hasNotificationFilters(filters.Include) && !matchesNotificationFilter(filters.Include, evt, identity, identityOK) {
		return false
	}
	if hasNotificationFilters(filters.Exclude) && matchesNotificationFilter(filters.Exclude, evt, identity, identityOK) {
		return false
	}
	return true
}

func hasNotificationFilters(filter NotificationFilter) bool {
	return len(filter.DeviceNames) > 0 || len(filter.Tags) > 0 || len(filter.IPs) > 0 || len(filter.Events) > 0
}

func hasIdentityFilters(filter NotificationFilter) bool {
	return len(filter.DeviceNames) > 0 || len(filter.Tags) > 0 || len(filter.IPs) > 0
}

func matchesNotificationFilter(filter NotificationFilter, evt event.Event, identity deviceIdentity, identityOK bool) bool {
	if len(filter.Events) > 0 && !matchesEventType(filter.Events, evt.EventType) {
		return false
	}
	if hasIdentityFilters(filter) && !identityOK {
		return false
	}
	if len(filter.DeviceNames) > 0 && !matchesFilterDeviceName(filter.DeviceNames, identity.Name) {
		return false
	}
	if len(filter.Tags) > 0 && !matchesSelectorAny(filter.Tags, identity.Tags) {
		return false
	}
	if len(filter.IPs) > 0 && !matchesFilterIPs(filter.IPs, identity.IPs) {
		return false
	}
	return true
}

func matchesFilterDeviceName(filters []string, value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	valueLower := strings.ToLower(value)
	for _, raw := range filters {
		pattern := strings.TrimSpace(raw)
		if pattern == "" {
			continue
		}
		if strings.ContainsAny(pattern, "*?[") {
			ok, err := path.Match(strings.ToLower(pattern), valueLower)
			if err != nil {
				continue
			}
			if ok {
				return true
			}
			continue
		}
		if strings.EqualFold(pattern, value) {
			return true
		}
	}
	return false
}

func matchesFilterIPs(filters []string, values []string) bool {
	if len(values) == 0 {
		return false
	}
	identityIPs := make([]netip.Addr, 0, len(values))
	for _, raw := range values {
		addr, err := netip.ParseAddr(strings.TrimSpace(raw))
		if err != nil {
			continue
		}
		identityIPs = append(identityIPs, addr)
	}
	if len(identityIPs) == 0 {
		return false
	}
	for _, raw := range filters {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		if addr, err := netip.ParseAddr(value); err == nil {
			for _, identityIP := range identityIPs {
				if identityIP == addr {
					return true
				}
			}
			continue
		}
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			continue
		}
		for _, identityIP := range identityIPs {
			if prefix.Contains(identityIP) {
				return true
			}
		}
	}
	return false
}

func matchesSelectorString(selector []string, value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	for _, raw := range selector {
		if strings.EqualFold(strings.TrimSpace(raw), value) {
			return true
		}
	}
	return false
}

func matchesSelectorAny(selector []string, values []string) bool {
	if len(values) == 0 {
		return false
	}
	for _, value := range values {
		if matchesSelectorString(selector, value) {
			return true
		}
	}
	return false
}

func matchesSelectorAnyIPs(selector []string, values []string) bool {
	if len(values) == 0 {
		return false
	}
	normalizedValues := make([]string, 0, len(values))
	for _, raw := range values {
		ip, err := netip.ParseAddr(strings.TrimSpace(raw))
		if err != nil {
			continue
		}
		normalizedValues = append(normalizedValues, ip.String())
	}
	if len(normalizedValues) == 0 {
		return false
	}
	for _, raw := range selector {
		ip, err := netip.ParseAddr(strings.TrimSpace(raw))
		if err != nil {
			continue
		}
		if contains(normalizedValues, ip.String()) {
			return true
		}
	}
	return false
}

func deviceIdentityFromEvent(evt event.Event) (deviceIdentity, bool) {
	if evt.SubjectType != event.SubjectPeer {
		return deviceIdentity{}, false
	}
	id := deviceIdentity{
		Name:   evt.SubjectID,
		Tags:   []string{},
		Owners: []string{},
		IPs:    []string{},
	}
	if evt.Payload == nil {
		return id, true
	}
	if rawName, ok := evt.Payload["name"]; ok {
		if name, ok := rawName.(string); ok && strings.TrimSpace(name) != "" {
			id.Name = strings.TrimSpace(name)
		}
	}
	id.Tags = payloadStringSlice(evt.Payload, "tags")
	id.Owners = payloadStringSlice(evt.Payload, "owners")
	id.IPs = payloadIPSlice(evt.Payload, "ips")
	return id, true
}

func payloadStringSlice(payload map[string]any, key string) []string {
	raw, ok := payload[key]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		out := make([]string, 0, len(v))
		for _, item := range v {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			out = append(out, item)
		}
		return out
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				continue
			}
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			out = append(out, s)
		}
		return out
	default:
		return nil
	}
}

func payloadIPSlice(payload map[string]any, key string) []string {
	values := payloadStringSlice(payload, key)
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		ip, err := netip.ParseAddr(value)
		if err != nil {
			continue
		}
		out = append(out, ip.String())
	}
	return out
}

func uniq(items []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

var ErrNoSinks = errors.New("no sinks configured")
