package logging

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	defer func() {
		os.Stdout = old
	}()

	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()

	fn()
	_ = w.Close()
	return <-done
}

func TestNewLoggerJSONIncludesLogSource(t *testing.T) {
	out := captureStdout(t, func() {
		logger, err := NewLogger(Config{Format: "json", Level: "info"})
		if err != nil {
			t.Fatalf("new logger: %v", err)
		}
		WithSource(logger, LogSourceSentinel).Info("json log", zap.String("custom", "value"))
		_ = logger.Sync()
	})

	line := strings.TrimSpace(out)
	if line == "" {
		t.Fatal("expected JSON log output")
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(line), &decoded); err != nil {
		t.Fatalf("decode JSON log: %v", err)
	}
	if got := decoded[LogSourceField]; got != LogSourceSentinel {
		t.Fatalf("expected log_source=%q, got %#v", LogSourceSentinel, got)
	}
	if got := decoded["message"]; got != "json log" {
		t.Fatalf("expected message json log, got %#v", got)
	}
}

func TestNewLoggerPrettyIncludesLogSource(t *testing.T) {
	out := captureStdout(t, func() {
		logger, err := NewLogger(Config{Format: "pretty", Level: "info", NoColor: true})
		if err != nil {
			t.Fatalf("new logger: %v", err)
		}
		WithSource(logger, LogSourceSentinel).Info("pretty log")
		_ = logger.Sync()
	})

	if !strings.Contains(out, LogSourceField) {
		t.Fatalf("expected pretty output to include %q: %q", LogSourceField, out)
	}
	if !strings.Contains(out, LogSourceSentinel) {
		t.Fatalf("expected pretty output to include %q: %q", LogSourceSentinel, out)
	}
}

func TestLogfAdapterEmitsStructuredRecord(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	logger := WithSource(zap.New(core), LogSourceTailscale)
	logf := LogfAdapter(logger, zapcore.InfoLevel)
	logf("  tailscale backend line  ")

	entries := observed.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}
	if got := entries[0].Message; got != "tailscale backend line" {
		t.Fatalf("expected trimmed message, got %q", got)
	}
	if got := entries[0].ContextMap()[LogSourceField]; got != LogSourceTailscale {
		t.Fatalf("expected log_source=%q, got %#v", LogSourceTailscale, got)
	}
}
