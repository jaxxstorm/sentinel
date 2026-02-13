package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadConfigYAMLWithEnvOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sentinel.yaml")
	content := "poll_interval: 30s\nstate:\n  path: file-from-config.json\noutput:\n  log_format: pretty\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SENTINEL_POLL_INTERVAL", "5s")
	t.Setenv("SENTINEL_STATE_PATH", filepath.Join(dir, "override-state.json"))

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PollInterval != 5*time.Second {
		t.Fatalf("expected poll interval from env override, got %s", cfg.PollInterval)
	}
	if cfg.State.Path != filepath.Join(dir, "override-state.json") {
		t.Fatalf("expected state path override, got %s", cfg.State.Path)
	}
}

func TestDefaultConfigHasStdoutDebugSink(t *testing.T) {
	cfg := Default()
	foundSink := false
	for _, sink := range cfg.Notifier.Sinks {
		if sink.Name == "stdout-debug" && sink.Type == "stdout" {
			foundSink = true
			break
		}
	}
	if !foundSink {
		t.Fatal("expected default stdout-debug sink")
	}
	if len(cfg.Notifier.Routes) == 0 {
		t.Fatal("expected at least one default notifier route")
	}
	foundRoute := false
	for _, route := range cfg.Notifier.Routes {
		for _, name := range route.Sinks {
			if name == "stdout-debug" {
				foundRoute = true
				break
			}
		}
	}
	if !foundRoute {
		t.Fatal("expected default route to include stdout-debug sink")
	}
}

func TestDefaultConfigSourceModeIsRealtime(t *testing.T) {
	cfg := Default()
	if cfg.Source.Mode != "realtime" {
		t.Fatalf("expected default source mode realtime, got %q", cfg.Source.Mode)
	}
}

func TestValidateSourceMode(t *testing.T) {
	cfg := Default()
	cfg.Source.Mode = "poll"
	if err := Validate(cfg); err != nil {
		t.Fatalf("expected poll mode to validate, got %v", err)
	}

	cfg.Source.Mode = "invalid-mode"
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation error for invalid source mode")
	}
	if !strings.Contains(err.Error(), "source.mode must be realtime or poll") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadExpandsNotifierSinkURLFromEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sentinel.yaml")
	content := "" +
		"notifier:\n" +
		"  sinks:\n" +
		"    - name: webhook-primary\n" +
		"      type: webhook\n" +
		"      url: ${REQUESTBIN_WEBHOOK_URL}\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("REQUESTBIN_WEBHOOK_URL", "https://example.com/hook")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Notifier.Sinks) == 0 {
		t.Fatal("expected at least one notifier sink")
	}
	if cfg.Notifier.Sinks[0].URL != "https://example.com/hook" {
		t.Fatalf("expected sink URL to be expanded from env, got %q", cfg.Notifier.Sinks[0].URL)
	}
}

func TestLoadExpandsMissingNotifierSinkURLToEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sentinel.yaml")
	content := "" +
		"notifier:\n" +
		"  sinks:\n" +
		"    - name: webhook-primary\n" +
		"      type: webhook\n" +
		"      url: ${MISSING_WEBHOOK_URL}\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Notifier.Sinks) == 0 {
		t.Fatal("expected at least one notifier sink")
	}
	if cfg.Notifier.Sinks[0].URL != "" {
		t.Fatalf("expected missing env expansion to produce empty string, got %q", cfg.Notifier.Sinks[0].URL)
	}
}

func TestValidateAcceptsWildcardNotifierRoute(t *testing.T) {
	cfg := Default()
	cfg.Notifier.Routes = []RouteConfig{{
		EventTypes: []string{"*"},
		Sinks:      []string{"stdout-debug"},
	}}
	if err := Validate(cfg); err != nil {
		t.Fatalf("expected wildcard route to validate, got %v", err)
	}
}

func TestValidateRejectsUnknownNotifierEventType(t *testing.T) {
	cfg := Default()
	cfg.Notifier.Routes = []RouteConfig{{
		EventTypes: []string{"peer.unknown"},
		Sinks:      []string{"stdout-debug"},
	}}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation error for unknown notifier event type")
	}
	if !strings.Contains(err.Error(), "unknown value") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRejectsEmptyNotifierEventTypes(t *testing.T) {
	cfg := Default()
	cfg.Notifier.Routes = []RouteConfig{{
		EventTypes: []string{},
		Sinks:      []string{"stdout-debug"},
	}}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation error for empty notifier event types")
	}
	if !strings.Contains(err.Error(), "must not be empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}
