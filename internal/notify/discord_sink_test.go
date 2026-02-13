package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jaxxstorm/sentinel/internal/event"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestDiscordSinkSendsExpectedPayloadAndLogsSuccess(t *testing.T) {
	var gotHeader string
	var gotPayload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Idempotency-Key")
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		gotPayload = body
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	core, logs := observer.New(zapcore.InfoLevel)
	sink := NewDiscordSink("discord-primary", srv.URL, zap.New(core))
	n := Notification{
		Event:          event.NewPeerEvent(event.TypePeerOnline, "peer1", "before", "after", map[string]any{"name": "node-a"}, time.Unix(1700000000, 0)),
		IdempotencyKey: "idempotency-1",
	}

	if err := sink.Send(context.Background(), n); err != nil {
		t.Fatal(err)
	}

	if gotHeader != "idempotency-1" {
		t.Fatalf("expected idempotency header idempotency-1, got %q", gotHeader)
	}
	embedsRaw, ok := gotPayload["embeds"].([]any)
	if !ok || len(embedsRaw) != 1 {
		t.Fatalf("expected 1 embed in payload, got %#v", gotPayload["embeds"])
	}
	embed, ok := embedsRaw[0].(map[string]any)
	if !ok {
		t.Fatalf("expected embed object, got %#v", embedsRaw[0])
	}
	if got := embed["title"]; got != "Sentinel peer.online" {
		t.Fatalf("expected title Sentinel peer.online, got %#v", got)
	}
	if got := embed["url"]; got != "https://login.tailscale.com/admin/machines" {
		t.Fatalf("expected embed url, got %#v", got)
	}
	if got := embed["color"]; got != float64(0x3498DB) {
		t.Fatalf("expected info color, got %#v", got)
	}
	fieldsRaw, ok := embed["fields"].([]any)
	if !ok || len(fieldsRaw) < 3 {
		t.Fatalf("expected at least 3 embed fields, got %#v", embed["fields"])
	}
	var payloadValue string
	for _, f := range fieldsRaw {
		field, ok := f.(map[string]any)
		if !ok {
			continue
		}
		if field["name"] == "Payload" {
			if v, ok := field["value"].(string); ok {
				payloadValue = v
			}
		}
	}
	if payloadValue == "" {
		t.Fatalf("expected payload field value, got %#v", fieldsRaw)
	}
	if payloadValue != "```json\n{\n  \"name\": \"node-a\"\n}\n```" {
		t.Fatalf("unexpected payload field value: %q", payloadValue)
	}

	entries := logs.FilterMessage("discord send succeeded").All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 success log, got %d", len(entries))
	}
	ctx := entries[0].ContextMap()
	if got := ctx["sink"]; got != "discord-primary" {
		t.Fatalf("expected sink discord-primary, got %#v", got)
	}
	if got := ctx["status_code"]; got != int64(http.StatusNoContent) {
		t.Fatalf("expected status_code %d, got %#v", http.StatusNoContent, got)
	}
}

func TestDiscordSinkPayloadFieldTruncation(t *testing.T) {
	payload := map[string]any{
		"blob": strings.Repeat("x", 5000),
	}
	n := Notification{
		Event:          event.NewPeerEvent(event.TypePeerHostinfoChanged, "peer1", "before", "after", payload, time.Unix(1700000000, 0)),
		IdempotencyKey: "idempotency-1",
	}
	out := discordWebhookPayload(n)
	if len(out.Embeds) != 1 || len(out.Embeds[0].Fields) < 3 {
		t.Fatalf("expected embed with payload field, got %#v", out)
	}
	payloadField := out.Embeds[0].Fields[2].Value
	if len(payloadField) > 1024 {
		t.Fatalf("expected payload field <= 1024 chars, got %d", len(payloadField))
	}
	if len(payloadField) < 3 || payloadField[len(payloadField)-3:] != "```" {
		t.Fatalf("expected payload field to remain fenced json block, got %q", payloadField)
	}
}

func TestDiscordSinkRetriesAndLogsFailures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	core, logs := observer.New(zapcore.WarnLevel)
	sink := NewDiscordSink("discord-primary", srv.URL, zap.New(core))
	sink.maxRetries = 1
	sink.backoff = time.Millisecond

	err := sink.Send(context.Background(), testNotification())
	if err == nil {
		t.Fatal("expected send error")
	}

	entries := logs.FilterMessage("discord send failed").All()
	if len(entries) != 2 {
		t.Fatalf("expected 2 failure logs for two attempts, got %d", len(entries))
	}
	ctx := entries[0].ContextMap()
	if got := ctx["sink"]; got != "discord-primary" {
		t.Fatalf("expected sink discord-primary, got %#v", got)
	}
	if got := ctx["status_code"]; got != int64(http.StatusBadGateway) {
		t.Fatalf("expected status_code %d, got %#v", http.StatusBadGateway, got)
	}
}
