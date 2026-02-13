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

func TestValidateAcceptsDiscordSink(t *testing.T) {
	cfg := Default()
	cfg.Notifier.Sinks = append(cfg.Notifier.Sinks, SinkConfig{
		Name: "discord-primary",
		Type: "discord",
		URL:  "https://discord.com/api/webhooks/a/b",
	})
	if err := Validate(cfg); err != nil {
		t.Fatalf("expected discord sink to validate, got %v", err)
	}
}

func TestValidateRejectsDiscordSinkWithoutURL(t *testing.T) {
	cfg := Default()
	cfg.Notifier.Sinks = []SinkConfig{{
		Name: "discord-primary",
		Type: "discord",
		URL:  "",
	}}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation error for discord sink without url")
	}
	if !strings.Contains(err.Error(), "url is required for discord sink") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRejectsUnsupportedSinkType(t *testing.T) {
	cfg := Default()
	cfg.Notifier.Sinks = []SinkConfig{{
		Name: "unknown",
		Type: "pagerduty",
		URL:  "https://example.com",
	}}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation error for unsupported sink type")
	}
	if !strings.Contains(err.Error(), "unsupported value") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadSupportsStructuredEnvOnlyConfig(t *testing.T) {
	t.Setenv(envVarDetectors, `{"presence":{"enabled":true},"runtime":{"enabled":false}}`)
	t.Setenv(envVarDetectorOrder, `["presence","runtime"]`)
	t.Setenv(envVarNotifierSinks, `[{"name":"stdout-debug","type":"stdout"},{"name":"discord-primary","type":"discord","url":"https://discord.com/api/webhooks/a/b"}]`)
	t.Setenv(envVarNotifierRoutes, `[{"event_types":["*"],"severities":[],"sinks":["stdout-debug","discord-primary"]}]`)
	t.Setenv(envVarStatePath, filepath.Join(t.TempDir(), "state.json"))

	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}

	if got := len(cfg.Detectors); got != 2 {
		t.Fatalf("expected 2 detectors from structured env, got %d", got)
	}
	if cfg.Detectors["runtime"].Enabled {
		t.Fatal("expected runtime detector to be disabled from structured env")
	}
	if len(cfg.DetectorOrder) != 2 || cfg.DetectorOrder[0] != "presence" || cfg.DetectorOrder[1] != "runtime" {
		t.Fatalf("unexpected detector order: %v", cfg.DetectorOrder)
	}
	if len(cfg.Notifier.Sinks) != 2 {
		t.Fatalf("expected 2 sinks from structured env, got %d", len(cfg.Notifier.Sinks))
	}
	if cfg.Notifier.Sinks[1].Type != "discord" {
		t.Fatalf("expected second sink to be discord, got %q", cfg.Notifier.Sinks[1].Type)
	}
	if len(cfg.Notifier.Routes) != 1 || len(cfg.Notifier.Routes[0].EventTypes) != 1 || cfg.Notifier.Routes[0].EventTypes[0] != "*" {
		t.Fatalf("unexpected notifier routes: %+v", cfg.Notifier.Routes)
	}
}

func TestLoadStructuredEnvOverridesTakePrecedence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sentinel.yaml")
	content := "" +
		"poll_interval: 30s\n" +
		"state:\n" +
		"  path: " + filepath.ToSlash(filepath.Join(dir, "state.json")) + "\n" +
		"notifier:\n" +
		"  sinks:\n" +
		"    - name: webhook-primary\n" +
		"      type: webhook\n" +
		"      url: https://example.invalid/original\n" +
		"  routes:\n" +
		"    - event_types: [\"peer.online\"]\n" +
		"      sinks: [\"webhook-primary\"]\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("SENTINEL_POLL_INTERVAL", "5s")
	t.Setenv(envVarNotifierSinks, `[{"name":"stdout-debug","type":"stdout"}]`)
	t.Setenv(envVarNotifierRoutes, `[{"event_types":["*"],"sinks":["stdout-debug"]}]`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PollInterval != 5*time.Second {
		t.Fatalf("expected scalar env override for poll interval, got %s", cfg.PollInterval)
	}
	if len(cfg.Notifier.Sinks) != 1 || cfg.Notifier.Sinks[0].Name != "stdout-debug" {
		t.Fatalf("expected structured env sinks to override file sinks, got %+v", cfg.Notifier.Sinks)
	}
	if len(cfg.Notifier.Routes) != 1 || cfg.Notifier.Routes[0].EventTypes[0] != "*" {
		t.Fatalf("expected structured env routes to override file routes, got %+v", cfg.Notifier.Routes)
	}
}

func TestLoadFailsOnMalformedStructuredEnvValue(t *testing.T) {
	t.Setenv(envVarNotifierSinks, "{bad")
	_, err := Load("")
	if err == nil {
		t.Fatal("expected parse error for malformed structured env value")
	}
	if !strings.Contains(err.Error(), envVarNotifierSinks) {
		t.Fatalf("expected error to include env key, got %v", err)
	}
}

func TestLoadFailsOnEmptyStructuredEnvValue(t *testing.T) {
	t.Setenv(envVarNotifierRoutes, "   ")
	_, err := Load("")
	if err == nil {
		t.Fatal("expected parse error for empty structured env value")
	}
	if !strings.Contains(err.Error(), envVarNotifierRoutes) {
		t.Fatalf("expected error to include env key, got %v", err)
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Fatalf("expected error to mention empty value, got %v", err)
	}
}

func TestLoadFileConfigRemainsUnchangedWithoutStructuredEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sentinel.yaml")
	content := "" +
		"state:\n" +
		"  path: " + filepath.ToSlash(filepath.Join(dir, "state.json")) + "\n" +
		"notifier:\n" +
		"  sinks:\n" +
		"    - name: webhook-primary\n" +
		"      type: webhook\n" +
		"      url: https://example.invalid/file\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, sink := range cfg.Notifier.Sinks {
		if sink.Name == "webhook-primary" && sink.URL == "https://example.invalid/file" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected file-based sink to remain present, got %+v", cfg.Notifier.Sinks)
	}
}
