package cli

import (
	"testing"
)

func TestRootCommandIncludesRequiredSubcommands(t *testing.T) {
	cmd := NewRootCommand()
	expected := []string{"run", "status", "diff", "dump-netmap", "test-notify", "validate-config", "version"}
	for _, name := range expected {
		found := false
		for _, c := range cmd.Commands() {
			if c.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected command %q", name)
		}
	}
}

func TestGlobalFlagsAvailableOnStatusCommand(t *testing.T) {
	cmd := NewRootCommand()
	status, _, err := cmd.Find([]string{"status"})
	if err != nil {
		t.Fatal(err)
	}
	for _, flag := range []string{
		"config", "log-format", "log-level", "no-color",
		"tailscale-auth-key", "tailscale-login-mode", "tailscale-state-dir", "tailscale-login-timeout", "tailscale-allow-interactive-fallback",
	} {
		if status.Flag(flag) == nil {
			t.Fatalf("expected flag %q on status command", flag)
		}
	}
}
