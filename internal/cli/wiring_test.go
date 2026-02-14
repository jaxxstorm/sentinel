package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jaxxstorm/sentinel/internal/source"
)

func TestBuildRuntimeAddsDefaultStdoutSinkWhenNotifierConfigEmpty(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "sentinel.yaml")
	cfg := "state:\n  path: " + filepath.ToSlash(filepath.Join(t.TempDir(), "state.json")) + "\nnotifier:\n  sinks: []\n  routes: []\n"
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}

	deps, err := buildRuntime(&GlobalOptions{ConfigPath: cfgPath})
	if err != nil {
		t.Fatal(err)
	}
	deps.runner.Source = source.NewStaticSource(source.Netmap{Peers: []source.Peer{{ID: "peer1", Name: "peer1", Online: true}}})
	deps.runner.Enrollment = nil

	res, err := deps.runner.RunOnce(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}
	if res.DryRunCount == 0 {
		t.Fatal("expected dry-run notification count > 0 with default stdout-debug sink")
	}
}

func TestBuildRuntimeFallsBackRouteToStdoutSinkWhenConfiguredSinksUnavailable(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "sentinel.yaml")
	cfg := "state:\n  path: " + filepath.ToSlash(filepath.Join(t.TempDir(), "state.json")) + "\nnotifier:\n  sinks:\n    - name: webhook-primary\n      type: webhook\n      url: \"${SLACK_WEBHOOK_URL}\"\n  routes:\n    - event_types: [\"peer.online\"]\n      sinks: [\"webhook-primary\"]\n"
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}

	deps, err := buildRuntime(&GlobalOptions{ConfigPath: cfgPath})
	if err != nil {
		t.Fatal(err)
	}
	deps.runner.Source = source.NewStaticSource(source.Netmap{Peers: []source.Peer{{ID: "peer2", Name: "peer2", Online: true}}})
	deps.runner.Enrollment = nil

	res, err := deps.runner.RunOnce(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}
	if res.DryRunCount == 0 {
		t.Fatal("expected dry-run notification count > 0 with route fallback to stdout-debug sink")
	}
}

func TestBuildRuntimeUsesRealtimeSourceByDefault(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "sentinel.yaml")
	cfg := "state:\n  path: " + filepath.ToSlash(filepath.Join(t.TempDir(), "state.json")) + "\n"
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}

	deps, err := buildRuntime(&GlobalOptions{ConfigPath: cfgPath})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := deps.source.(*source.TSNetRealtimeSource); !ok {
		t.Fatalf("expected realtime source by default, got %T", deps.source)
	}
}

func TestBuildRuntimeUsesPollingSourceWhenConfigured(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "sentinel.yaml")
	cfg := "state:\n  path: " + filepath.ToSlash(filepath.Join(t.TempDir(), "state.json")) + "\nsource:\n  mode: poll\n"
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}

	deps, err := buildRuntime(&GlobalOptions{ConfigPath: cfgPath})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := deps.source.(*source.TSNetSource); !ok {
		t.Fatalf("expected polling source when source.mode=poll, got %T", deps.source)
	}
}

func TestBuildRuntimeDeliversEventToDiscordSink(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	cfgPath := filepath.Join(t.TempDir(), "sentinel.yaml")
	cfg := "state:\n  path: " + filepath.ToSlash(filepath.Join(t.TempDir(), "state.json")) + "\n" +
		"notifier:\n" +
		"  sinks:\n" +
		"    - name: discord-primary\n" +
		"      type: discord\n" +
		"      url: \"" + srv.URL + "\"\n" +
		"  routes:\n" +
		"    - event_types: [\"peer.online\"]\n" +
		"      sinks: [\"discord-primary\"]\n"
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}

	deps, err := buildRuntime(&GlobalOptions{ConfigPath: cfgPath})
	if err != nil {
		t.Fatal(err)
	}
	deps.runner.Source = source.NewStaticSource(source.Netmap{Peers: []source.Peer{{ID: "peer-discord", Name: "peer-discord", Online: true}}})
	deps.runner.Enrollment = nil

	if _, err := deps.runner.RunOnce(context.Background(), false); err != nil {
		t.Fatal(err)
	}
	if requests != 1 {
		t.Fatalf("expected 1 discord request, got %d", requests)
	}
}

func TestBuildRuntimeSupportsEnvOnlyConfig(t *testing.T) {
	t.Setenv("SENTINEL_STATE_PATH", filepath.ToSlash(filepath.Join(t.TempDir(), "state.json")))
	t.Setenv("SENTINEL_NOTIFIER_SINKS", `[{"name":"stdout-debug","type":"stdout"}]`)
	t.Setenv("SENTINEL_NOTIFIER_ROUTES", `[{"event_types":["peer.online"],"sinks":["stdout-debug"]}]`)

	deps, err := buildRuntime(&GlobalOptions{})
	if err != nil {
		t.Fatal(err)
	}
	deps.runner.Source = source.NewStaticSource(source.Netmap{Peers: []source.Peer{{ID: "peer-env", Name: "peer-env", Online: true}}})
	deps.runner.Enrollment = nil

	res, err := deps.runner.RunOnce(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}
	if res.DryRunCount == 0 {
		t.Fatal("expected dry-run notification count > 0 for env-only config")
	}
}

func TestBuildRuntimeResolvesOAuthCredentialsAndTagsFromEnv(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "sentinel.yaml")
	cfg := "state:\n  path: " + filepath.ToSlash(filepath.Join(t.TempDir(), "state.json")) + "\n"
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("SENTINEL_TSNET_CLIENT_SECRET", "env-secret")
	t.Setenv("SENTINEL_TSNET_CLIENT_ID", "env-client-id")
	t.Setenv("SENTINEL_TSNET_ADVERTISE_TAGS", `["tag:sentinel"]`)

	deps, err := buildRuntime(&GlobalOptions{ConfigPath: cfgPath})
	if err != nil {
		t.Fatal(err)
	}
	if deps.cfg.TSNet.ClientSecret != "env-secret" {
		t.Fatalf("expected client secret from env, got %q", deps.cfg.TSNet.ClientSecret)
	}
	if deps.cfg.TSNet.ClientID != "env-client-id" {
		t.Fatalf("expected client id from env, got %q", deps.cfg.TSNet.ClientID)
	}
	if len(deps.cfg.TSNet.AdvertiseTags) != 1 || deps.cfg.TSNet.AdvertiseTags[0] != "tag:sentinel" {
		t.Fatalf("unexpected advertise tags: %v", deps.cfg.TSNet.AdvertiseTags)
	}
	if deps.cfg.TSNet.CredentialMode != "oauth" {
		t.Fatalf("expected oauth credential mode, got %q", deps.cfg.TSNet.CredentialMode)
	}
}
