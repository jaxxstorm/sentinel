package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jaxxstorm/sentinel/internal/event"
	"github.com/spf13/viper"
)

type Config struct {
	PollInterval   time.Duration       `mapstructure:"poll_interval" json:"poll_interval"`
	PollJitter     time.Duration       `mapstructure:"poll_jitter" json:"poll_jitter"`
	PollBackoffMin time.Duration       `mapstructure:"poll_backoff_min" json:"poll_backoff_min"`
	PollBackoffMax time.Duration       `mapstructure:"poll_backoff_max" json:"poll_backoff_max"`
	Source         SourceConfig        `mapstructure:"source" json:"source"`
	Detectors      map[string]Detector `mapstructure:"detectors" json:"detectors"`
	DetectorOrder  []string            `mapstructure:"detector_order" json:"detector_order"`
	Policy         PolicyConfig        `mapstructure:"policy" json:"policy"`
	Notifier       NotifierConfig      `mapstructure:"notifier" json:"notifier"`
	State          StateConfig         `mapstructure:"state" json:"state"`
	Output         OutputConfig        `mapstructure:"output" json:"output"`
	TSNet          TSNetConfig         `mapstructure:"tsnet" json:"tsnet"`
}

type Detector struct {
	Enabled bool `mapstructure:"enabled" json:"enabled"`
}

type SourceConfig struct {
	Mode string `mapstructure:"mode" json:"mode"`
}

type PolicyConfig struct {
	DebounceWindow    time.Duration `mapstructure:"debounce_window" json:"debounce_window"`
	SuppressionWindow time.Duration `mapstructure:"suppression_window" json:"suppression_window"`
	RateLimitPerMin   int           `mapstructure:"rate_limit_per_min" json:"rate_limit_per_min"`
	BatchSize         int           `mapstructure:"batch_size" json:"batch_size"`
}

type NotifierConfig struct {
	IdempotencyKeyTTL time.Duration `mapstructure:"idempotency_key_ttl" json:"idempotency_key_ttl"`
	Routes            []RouteConfig `mapstructure:"routes" json:"routes"`
	Sinks             []SinkConfig  `mapstructure:"sinks" json:"sinks"`
}

type RouteConfig struct {
	EventTypes []string `mapstructure:"event_types" json:"event_types"`
	Severities []string `mapstructure:"severities" json:"severities"`
	Sinks      []string `mapstructure:"sinks" json:"sinks"`
}

type SinkConfig struct {
	Name string `mapstructure:"name" json:"name"`
	Type string `mapstructure:"type" json:"type"`
	URL  string `mapstructure:"url" json:"url"`
}

type StateConfig struct {
	Path              string        `mapstructure:"path" json:"path"`
	IdempotencyKeyTTL time.Duration `mapstructure:"idempotency_key_ttl" json:"idempotency_key_ttl"`
}

type OutputConfig struct {
	LogFormat string `mapstructure:"log_format" json:"log_format"`
	LogLevel  string `mapstructure:"log_level" json:"log_level"`
	NoColor   bool   `mapstructure:"no_color" json:"no_color"`
}

type TSNetConfig struct {
	Hostname                 string        `mapstructure:"hostname" json:"hostname"`
	StateDir                 string        `mapstructure:"state_dir" json:"state_dir"`
	AuthKey                  string        `mapstructure:"auth_key" json:"auth_key"`
	AdvertiseTags            []string      `mapstructure:"advertise_tags" json:"advertise_tags"`
	ClientSecret             string        `mapstructure:"client_secret" json:"client_secret"`
	ClientID                 string        `mapstructure:"client_id" json:"client_id"`
	IDToken                  string        `mapstructure:"id_token" json:"id_token"`
	Audience                 string        `mapstructure:"audience" json:"audience"`
	LoginMode                string        `mapstructure:"login_mode" json:"login_mode"`
	AllowInteractiveFallback bool          `mapstructure:"allow_interactive_fallback" json:"allow_interactive_fallback"`
	LoginTimeout             time.Duration `mapstructure:"login_timeout" json:"login_timeout"`
	AuthKeySource            string        `mapstructure:"-"`
	OAuthSource              string        `mapstructure:"-"`
	CredentialMode           string        `mapstructure:"-"`
	CredentialSource         string        `mapstructure:"-"`
}

const (
	envVarConfigPath        = "SENTINEL_CONFIG_PATH"
	envVarStatePath         = "SENTINEL_STATE_PATH"
	envVarTSNetTags         = "SENTINEL_TSNET_ADVERTISE_TAGS"
	envVarTSNetClientSecret = "SENTINEL_TSNET_CLIENT_SECRET"
	envVarTSNetClientID     = "SENTINEL_TSNET_CLIENT_ID"
	envVarTSNetIDToken      = "SENTINEL_TSNET_ID_TOKEN"
	envVarTSNetAudience     = "SENTINEL_TSNET_AUDIENCE"
	envVarDetectors         = "SENTINEL_DETECTORS"
	envVarDetectorOrder     = "SENTINEL_DETECTOR_ORDER"
	envVarNotifierSinks     = "SENTINEL_NOTIFIER_SINKS"
	envVarNotifierRoutes    = "SENTINEL_NOTIFIER_ROUTES"
)

var advertiseTagPattern = regexp.MustCompile(`^tag:[A-Za-z0-9._-]+$`)

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
	envConfigPath := strings.TrimSpace(os.Getenv(envVarConfigPath))

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
	suppressStructuredEnvForViper(v)

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal config: %w", err)
	}
	if err := applyStructuredEnvOverrides(&cfg); err != nil {
		return cfg, err
	}
	expandEnvPlaceholders(&cfg)
	if envPath := strings.TrimSpace(os.Getenv(envVarStatePath)); envPath != "" {
		cfg.State.Path = envPath
	}
	if rawTags, ok := os.LookupEnv(envVarTSNetTags); ok {
		tags, err := parseStringListEnv(envVarTSNetTags, rawTags)
		if err != nil {
			return cfg, err
		}
		cfg.TSNet.AdvertiseTags = tags
	}
	applyStringEnvOverride(envVarTSNetClientSecret, &cfg.TSNet.ClientSecret)
	applyStringEnvOverride(envVarTSNetClientID, &cfg.TSNet.ClientID)
	applyStringEnvOverride(envVarTSNetIDToken, &cfg.TSNet.IDToken)
	applyStringEnvOverride(envVarTSNetAudience, &cfg.TSNet.Audience)
	return cfg, Validate(cfg)
}

func applyStructuredEnvOverrides(cfg *Config) error {
	var detectors map[string]Detector
	if present, err := decodeEnvJSON(envVarDetectors, &detectors); err != nil {
		return err
	} else if present {
		cfg.Detectors = detectors
		if cfg.Detectors == nil {
			cfg.Detectors = map[string]Detector{}
		}
	}
	var detectorOrder []string
	if present, err := decodeEnvJSON(envVarDetectorOrder, &detectorOrder); err != nil {
		return err
	} else if present {
		cfg.DetectorOrder = detectorOrder
	}
	var sinks []SinkConfig
	if present, err := decodeEnvJSON(envVarNotifierSinks, &sinks); err != nil {
		return err
	} else if present {
		cfg.Notifier.Sinks = sinks
	}
	var routes []RouteConfig
	if present, err := decodeEnvJSON(envVarNotifierRoutes, &routes); err != nil {
		return err
	} else if present {
		cfg.Notifier.Routes = routes
	}
	return nil
}

func suppressStructuredEnvForViper(v *viper.Viper) {
	if _, ok := os.LookupEnv(envVarDetectors); ok {
		v.Set("detectors", map[string]Detector{})
	}
	if _, ok := os.LookupEnv(envVarDetectorOrder); ok {
		v.Set("detector_order", []string{})
	}
	if _, ok := os.LookupEnv(envVarNotifierSinks); ok {
		v.Set("notifier.sinks", []SinkConfig{})
	}
	if _, ok := os.LookupEnv(envVarNotifierRoutes); ok {
		v.Set("notifier.routes", []RouteConfig{})
	}
}

func decodeEnvJSON(key string, target any) (bool, error) {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return false, nil
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return true, fmt.Errorf("parse %s: value is empty; expected JSON", key)
	}
	if err := json.Unmarshal([]byte(raw), target); err != nil {
		return true, fmt.Errorf("parse %s: %w", key, err)
	}
	return true, nil
}

func parseStringListEnv(key, raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("parse %s: value is empty; expected JSON array or comma-separated list", key)
	}
	if strings.HasPrefix(raw, "[") {
		var values []string
		if err := json.Unmarshal([]byte(raw), &values); err != nil {
			return nil, fmt.Errorf("parse %s: %w", key, err)
		}
		for i, value := range values {
			values[i] = strings.TrimSpace(value)
			if values[i] == "" {
				return nil, fmt.Errorf("parse %s: values must not be empty", key)
			}
		}
		return values, nil
	}
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			return nil, fmt.Errorf("parse %s: values must not be empty", key)
		}
		values = append(values, value)
	}
	return values, nil
}

func applyStringEnvOverride(key string, target *string) {
	if raw, ok := os.LookupEnv(key); ok {
		*target = strings.TrimSpace(raw)
	}
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
	case "", "auto", "auth_key", "oauth", "interactive":
	default:
		return fmt.Errorf("tsnet.login_mode must be auto, auth_key, oauth, or interactive")
	}
	if cfg.TSNet.LoginTimeout <= 0 {
		return fmt.Errorf("tsnet.login_timeout must be > 0")
	}
	for i, rawTag := range cfg.TSNet.AdvertiseTags {
		tag := strings.TrimSpace(rawTag)
		if tag == "" {
			return fmt.Errorf("tsnet.advertise_tags[%d] must not be empty", i)
		}
		if !advertiseTagPattern.MatchString(tag) {
			return fmt.Errorf("tsnet.advertise_tags[%d] must match tag:<name> format", i)
		}
		cfg.TSNet.AdvertiseTags[i] = tag
	}
	clientSecret := strings.TrimSpace(cfg.TSNet.ClientSecret)
	clientID := strings.TrimSpace(cfg.TSNet.ClientID)
	idToken := strings.TrimSpace(cfg.TSNet.IDToken)
	audience := strings.TrimSpace(cfg.TSNet.Audience)
	cfg.TSNet.ClientSecret = clientSecret
	cfg.TSNet.ClientID = clientID
	cfg.TSNet.IDToken = idToken
	cfg.TSNet.Audience = audience
	if mode == "oauth" && clientSecret == "" {
		return fmt.Errorf("tsnet.client_secret is required for oauth login mode")
	}
	if clientSecret == "" && (clientID != "" || idToken != "" || audience != "") {
		return fmt.Errorf("tsnet.client_secret is required when oauth credential fields are set")
	}
	if clientSecret != "" && clientID == "" {
		return fmt.Errorf("tsnet.client_id is required when tsnet.client_secret is set")
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
