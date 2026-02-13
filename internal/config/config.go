package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jaxxstorm/sentinel/internal/event"
	"github.com/spf13/viper"
)

type Config struct {
	PollInterval   time.Duration       `mapstructure:"poll_interval"`
	PollJitter     time.Duration       `mapstructure:"poll_jitter"`
	PollBackoffMin time.Duration       `mapstructure:"poll_backoff_min"`
	PollBackoffMax time.Duration       `mapstructure:"poll_backoff_max"`
	Source         SourceConfig        `mapstructure:"source"`
	Detectors      map[string]Detector `mapstructure:"detectors"`
	DetectorOrder  []string            `mapstructure:"detector_order"`
	Policy         PolicyConfig        `mapstructure:"policy"`
	Notifier       NotifierConfig      `mapstructure:"notifier"`
	State          StateConfig         `mapstructure:"state"`
	Output         OutputConfig        `mapstructure:"output"`
	TSNet          TSNetConfig         `mapstructure:"tsnet"`
}

type Detector struct {
	Enabled bool `mapstructure:"enabled"`
}

type SourceConfig struct {
	Mode string `mapstructure:"mode"`
}

type PolicyConfig struct {
	DebounceWindow    time.Duration `mapstructure:"debounce_window"`
	SuppressionWindow time.Duration `mapstructure:"suppression_window"`
	RateLimitPerMin   int           `mapstructure:"rate_limit_per_min"`
	BatchSize         int           `mapstructure:"batch_size"`
}

type NotifierConfig struct {
	IdempotencyKeyTTL time.Duration `mapstructure:"idempotency_key_ttl"`
	Routes            []RouteConfig `mapstructure:"routes"`
	Sinks             []SinkConfig  `mapstructure:"sinks"`
}

type RouteConfig struct {
	EventTypes []string `mapstructure:"event_types"`
	Severities []string `mapstructure:"severities"`
	Sinks      []string `mapstructure:"sinks"`
}

type SinkConfig struct {
	Name string `mapstructure:"name"`
	Type string `mapstructure:"type"`
	URL  string `mapstructure:"url"`
}

type StateConfig struct {
	Path              string        `mapstructure:"path"`
	IdempotencyKeyTTL time.Duration `mapstructure:"idempotency_key_ttl"`
}

type OutputConfig struct {
	LogFormat string `mapstructure:"log_format"`
	LogLevel  string `mapstructure:"log_level"`
	NoColor   bool   `mapstructure:"no_color"`
}

type TSNetConfig struct {
	Hostname                 string        `mapstructure:"hostname"`
	StateDir                 string        `mapstructure:"state_dir"`
	AuthKey                  string        `mapstructure:"auth_key"`
	LoginMode                string        `mapstructure:"login_mode"`
	AllowInteractiveFallback bool          `mapstructure:"allow_interactive_fallback"`
	LoginTimeout             time.Duration `mapstructure:"login_timeout"`
	AuthKeySource            string        `mapstructure:"-"`
}

func Default() Config {
	return Config{
		PollInterval:   10 * time.Second,
		PollJitter:     1 * time.Second,
		PollBackoffMin: 500 * time.Millisecond,
		PollBackoffMax: 30 * time.Second,
		Source: SourceConfig{
			Mode: "realtime",
		},
		Detectors: map[string]Detector{
			"presence":     {Enabled: true},
			"peer_changes": {Enabled: true},
			"runtime":      {Enabled: true},
		},
		DetectorOrder: []string{"presence", "peer_changes", "runtime"},
		Policy: PolicyConfig{
			DebounceWindow:    3 * time.Second,
			SuppressionWindow: 0,
			RateLimitPerMin:   120,
			BatchSize:         20,
		},
		Notifier: NotifierConfig{
			IdempotencyKeyTTL: 24 * time.Hour,
			Routes: []RouteConfig{{
				EventTypes: []string{"*"},
				Sinks:      []string{"stdout-debug"},
			}},
			Sinks: []SinkConfig{
				{Name: "stdout-debug", Type: "stdout"},
				{Name: "webhook-primary", Type: "webhook", URL: "${SLACK_WEBHOOK_URL}"},
			},
		},
		State: StateConfig{
			Path:              ".sentinel/state.json",
			IdempotencyKeyTTL: 24 * time.Hour,
		},
		Output: OutputConfig{LogFormat: "pretty", LogLevel: "info", NoColor: false},
		TSNet: TSNetConfig{
			Hostname:                 "sentinel",
			StateDir:                 ".sentinel/tsnet",
			LoginMode:                "auto",
			AllowInteractiveFallback: false,
			LoginTimeout:             5 * time.Minute,
		},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	v := viper.New()
	v.SetEnvPrefix("SENTINEL")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	// Check env config too before checking file path
	envConfigPath := v.GetString("SENTINEL_CONFIG_PATH")

	if path == "" && envConfigPath != "" {
		path = envConfigPath
	}

	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return cfg, fmt.Errorf("read config: %w", err)
		}
	} else {
		if _, err := os.Stat("sentinel.yaml"); err == nil {
			v.SetConfigFile("sentinel.yaml")
			_ = v.ReadInConfig()
		} else if _, err := os.Stat("sentinel.json"); err == nil {
			v.SetConfigFile("sentinel.json")
			_ = v.ReadInConfig()
		}
	}

	v.SetDefault("poll_interval", cfg.PollInterval)
	v.SetDefault("poll_jitter", cfg.PollJitter)
	v.SetDefault("poll_backoff_min", cfg.PollBackoffMin)
	v.SetDefault("poll_backoff_max", cfg.PollBackoffMax)
	v.SetDefault("source.mode", cfg.Source.Mode)
	v.SetDefault("detector_order", cfg.DetectorOrder)
	v.SetDefault("output.log_format", cfg.Output.LogFormat)
	v.SetDefault("output.log_level", cfg.Output.LogLevel)
	v.SetDefault("state.path", cfg.State.Path)
	v.SetDefault("tsnet.hostname", cfg.TSNet.Hostname)
	v.SetDefault("tsnet.state_dir", cfg.TSNet.StateDir)
	v.SetDefault("tsnet.login_mode", cfg.TSNet.LoginMode)
	v.SetDefault("tsnet.allow_interactive_fallback", cfg.TSNet.AllowInteractiveFallback)
	v.SetDefault("tsnet.login_timeout", cfg.TSNet.LoginTimeout)

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal config: %w", err)
	}
	expandEnvPlaceholders(&cfg)
	if envPath := os.Getenv("SENTINEL_STATE_PATH"); envPath != "" {
		cfg.State.Path = envPath
	}
	return cfg, Validate(cfg)
}

func expandEnvPlaceholders(cfg *Config) {
	for i := range cfg.Notifier.Sinks {
		url := strings.TrimSpace(cfg.Notifier.Sinks[i].URL)
		if strings.Contains(url, "${") {
			cfg.Notifier.Sinks[i].URL = os.ExpandEnv(url)
		}
	}
}

func Validate(cfg Config) error {
	if cfg.PollInterval <= 0 {
		return fmt.Errorf("poll_interval must be > 0")
	}
	if cfg.Policy.BatchSize <= 0 {
		return fmt.Errorf("policy.batch_size must be > 0")
	}
	if len(cfg.DetectorOrder) == 0 {
		return fmt.Errorf("detector_order must not be empty")
	}
	for _, name := range cfg.DetectorOrder {
		if _, ok := cfg.Detectors[name]; !ok {
			return fmt.Errorf("detector_order references unknown detector %q", name)
		}
	}
	if cfg.State.Path == "" {
		return fmt.Errorf("state.path is required")
	}
	if !filepath.IsAbs(cfg.State.Path) {
		cfg.State.Path = filepath.Clean(cfg.State.Path)
	}
	if !strings.EqualFold(cfg.Output.LogFormat, "pretty") && !strings.EqualFold(cfg.Output.LogFormat, "json") {
		return fmt.Errorf("output.log_format must be pretty or json")
	}
	if cfg.TSNet.StateDir == "" {
		return fmt.Errorf("tsnet.state_dir is required")
	}
	mode := strings.ToLower(strings.TrimSpace(cfg.TSNet.LoginMode))
	switch mode {
	case "", "auto", "auth_key", "interactive":
	default:
		return fmt.Errorf("tsnet.login_mode must be auto, auth_key, or interactive")
	}
	if cfg.TSNet.LoginTimeout <= 0 {
		return fmt.Errorf("tsnet.login_timeout must be > 0")
	}
	sourceMode := strings.ToLower(strings.TrimSpace(cfg.Source.Mode))
	switch sourceMode {
	case "", "realtime", "poll":
	default:
		return fmt.Errorf("source.mode must be realtime or poll")
	}
	for i, route := range cfg.Notifier.Routes {
		if len(route.EventTypes) == 0 {
			return fmt.Errorf("notifier.routes[%d].event_types must not be empty", i)
		}
		for j, et := range route.EventTypes {
			et = strings.TrimSpace(et)
			if et == "" {
				return fmt.Errorf("notifier.routes[%d].event_types[%d] must not be empty", i, j)
			}
			if et != "*" && !event.IsKnownType(et) {
				return fmt.Errorf("notifier.routes[%d].event_types[%d] has unknown value %q", i, j, et)
			}
		}
	}
	for i, sink := range cfg.Notifier.Sinks {
		sinkType := strings.ToLower(strings.TrimSpace(sink.Type))
		switch sinkType {
		case "", "webhook", "stdout", "debug", "discord":
		default:
			return fmt.Errorf("notifier.sinks[%d].type has unsupported value %q", i, sink.Type)
		}
		if sinkType == "discord" && strings.TrimSpace(sink.URL) == "" {
			return fmt.Errorf("notifier.sinks[%d].url is required for discord sink", i)
		}
	}
	return nil
}
