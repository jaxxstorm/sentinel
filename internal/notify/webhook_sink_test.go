package notify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jaxxstorm/sentinel/internal/event"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func testNotification() Notification {
	return Notification{
		Event:          event.NewPresenceEvent(event.TypePeerOnline, "peer1", "before", "after", nil, time.Now()),
		IdempotencyKey: "k1",
	}
}

func TestWebhookSinkLogsSuccessWithStatusCodeAndSinkName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	core, logs := observer.New(zapcore.InfoLevel)
	sink := NewWebhookSink("webhook-primary", srv.URL, zap.New(core))

	if err := sink.Send(context.Background(), testNotification()); err != nil {
		t.Fatal(err)
	}

	entries := logs.FilterMessage("webhook send succeeded").All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 success log, got %d", len(entries))
	}
	ctx := entries[0].ContextMap()
	if got := ctx["sink"]; got != "webhook-primary" {
		t.Fatalf("expected sink webhook-primary, got %#v", got)
	}
	if got := ctx["status_code"]; got != int64(http.StatusNoContent) {
		t.Fatalf("expected status_code %d, got %#v", http.StatusNoContent, got)
	}
}

func TestWebhookSinkLogsFailuresWithStatusCodeAndSinkName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	core, logs := observer.New(zapcore.WarnLevel)
	sink := NewWebhookSink("webhook-primary", srv.URL, zap.New(core))
	sink.maxRetries = 1
	sink.backoff = time.Millisecond

	err := sink.Send(context.Background(), testNotification())
	if err == nil {
		t.Fatal("expected send error")
	}

	entries := logs.FilterMessage("webhook send failed").All()
	if len(entries) == 0 {
		t.Fatal("expected webhook send failed logs")
	}
	ctx := entries[0].ContextMap()
	if got := ctx["sink"]; got != "webhook-primary" {
		t.Fatalf("expected sink webhook-primary, got %#v", got)
	}
	if got := ctx["status_code"]; got != int64(http.StatusBadGateway) {
		t.Fatalf("expected status_code %d, got %#v", http.StatusBadGateway, got)
	}
}
