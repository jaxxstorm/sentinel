package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jaxxstorm/sentinel/internal/app"
	"github.com/jaxxstorm/sentinel/internal/config"
	"github.com/jaxxstorm/sentinel/internal/diff"
	"github.com/jaxxstorm/sentinel/internal/logging"
	"github.com/jaxxstorm/sentinel/internal/metrics"
	"github.com/jaxxstorm/sentinel/internal/notify"
	"github.com/jaxxstorm/sentinel/internal/onboarding"
	"github.com/jaxxstorm/sentinel/internal/output"
	"github.com/jaxxstorm/sentinel/internal/policy"
	"github.com/jaxxstorm/sentinel/internal/source"
	"github.com/jaxxstorm/sentinel/internal/state"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"tailscale.com/tsnet"
)

type runtimeDeps struct {
	cfg        config.Config
	runner     *app.Runner
	renderer   *output.Renderer
	source     source.NetmapSource
	notifier   *notify.Notifier
	enrollment onboarding.EnrollmentManager
}

func buildRuntime(opts *GlobalOptions) (*runtimeDeps, error) {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return nil, err
	}
	if opts.LogFormat != "" {
		cfg.Output.LogFormat = opts.LogFormat
	}
	if opts.LogLevel != "" {
		cfg.Output.LogLevel = opts.LogLevel
	}
	if opts.NoColor {
		cfg.Output.NoColor = true
	}
	if opts.TailscaleLoginMode != "" {
		cfg.TSNet.LoginMode = strings.ToLower(strings.TrimSpace(opts.TailscaleLoginMode))
	}
	if opts.TailscaleStateDir != "" {
		cfg.TSNet.StateDir = opts.TailscaleStateDir
	}
	if opts.TailscaleLoginTimeout > 0 {
		cfg.TSNet.LoginTimeout = opts.TailscaleLoginTimeout
	}
	if opts.TailscaleFallbackOverride {
		cfg.TSNet.AllowInteractiveFallback = true
	}
	authKey, sourceName := onboarding.ResolveAuthKey(
		opts.TailscaleAuthKey,
		os.Getenv("SENTINEL_TAILSCALE_AUTH_KEY"),
		cfg.TSNet.AuthKey,
	)
	cfg.TSNet.AuthKey = authKey
	cfg.TSNet.AuthKeySource = sourceName
	if err := config.Validate(cfg); err != nil {
		return nil, err
	}

	logger, err := logging.NewLogger(logging.Config{
		Format:  cfg.Output.LogFormat,
		Level:   cfg.Output.LogLevel,
		NoColor: cfg.Output.NoColor,
	})
	if err != nil {
		return nil, err
	}
	sentinelLogger := logging.WithSource(logger, logging.LogSourceSentinel)
	st := state.NewFileStore(cfg.State.Path)
	detectors := []diff.Detector{
		diff.NewPresenceDetector(),
		diff.NewPeerChangeDetector(),
		diff.NewRuntimeDetector(),
	}
	engine := diff.NewEngine(detectors)
	policyEngine := policy.NewEngine(policy.Config{
		DebounceWindow:    cfg.Policy.DebounceWindow,
		SuppressionWindow: cfg.Policy.SuppressionWindow,
		RateLimitPerMin:   cfg.Policy.RateLimitPerMin,
		BatchSize:         cfg.Policy.BatchSize,
	})

	const defaultSinkName = "stdout-debug"
	sinks := make([]notify.Sink, 0, len(cfg.Notifier.Sinks))
	availableSinks := map[string]struct{}{}
	defaultSinkPresent := false
	for _, sinkCfg := range cfg.Notifier.Sinks {
		sinkType := strings.ToLower(strings.TrimSpace(sinkCfg.Type))
		switch sinkType {
		case "", "webhook":
			url := strings.TrimSpace(sinkCfg.URL)
			if url == "" || strings.Contains(url, "${") {
				sentinelLogger.Warn("skipping sink with empty/unresolved URL", zap.String("sink", sinkCfg.Name))
				continue
			}
			sink := notify.NewWebhookSink(sinkCfg.Name, url, logging.WithSource(logger, logging.LogSourceSink))
			sinks = append(sinks, sink)
			availableSinks[sink.Name()] = struct{}{}
		case "stdout", "debug":
			name := sinkCfg.Name
			if name == "" {
				name = defaultSinkName
			}
			sink := notify.NewStdoutSink(name, os.Stdout)
			sinks = append(sinks, sink)
			availableSinks[sink.Name()] = struct{}{}
			if sink.Name() == defaultSinkName {
				defaultSinkPresent = true
			}
		case "discord":
			url := strings.TrimSpace(sinkCfg.URL)
			if url == "" || strings.Contains(url, "${") {
				return nil, fmt.Errorf("discord sink %q requires a non-empty webhook url", sinkCfg.Name)
			}
			sink := notify.NewDiscordSink(sinkCfg.Name, url, logging.WithSource(logger, logging.LogSourceSink))
			sinks = append(sinks, sink)
			availableSinks[sink.Name()] = struct{}{}
		default:
			return nil, fmt.Errorf("unsupported sink type %q", sinkCfg.Type)
		}
	}
	if !defaultSinkPresent {
		sentinelLogger.Info("adding default stdout-debug sink")
		sink := notify.NewStdoutSink(defaultSinkName, os.Stdout)
		sinks = append(sinks, sink)
		availableSinks[sink.Name()] = struct{}{}
	}

	routes := make([]notify.Route, 0, len(cfg.Notifier.Routes))
	for _, r := range cfg.Notifier.Routes {
		validSinks := make([]string, 0, len(r.Sinks))
		for _, sinkName := range r.Sinks {
			if _, ok := availableSinks[sinkName]; ok {
				validSinks = append(validSinks, sinkName)
			}
		}
		if len(validSinks) == 0 {
			sentinelLogger.Warn("route has no available sinks; falling back to stdout-debug", zap.Strings("event_types", r.EventTypes))
			validSinks = []string{defaultSinkName}
		}
		routes = append(routes, notify.Route{EventTypes: r.EventTypes, Severities: r.Severities, Sinks: validSinks})
	}
	if len(routes) == 0 {
		sentinelLogger.Info("adding default notifier route to stdout-debug")
		routes = append(routes, notify.Route{
			EventTypes: []string{"*"},
			Sinks:      []string{defaultSinkName},
		})
	}
	notifier := notify.New(notify.Config{Routes: routes, IdempotencyKeyTTL: cfg.Notifier.IdempotencyKeyTTL}, st, sinks)

	ts := &tsnet.Server{
		Hostname: cfg.TSNet.Hostname,
		Dir:      cfg.TSNet.StateDir,
		UserLogf: logging.LogfAdapter(logging.WithSource(logger, logging.LogSourceTailscale), zapcore.InfoLevel),
		Logf:     logging.LogfAdapter(logging.WithSource(logger, logging.LogSourceTailscale), zapcore.DebugLevel),
	}
	var src source.NetmapSource
	switch strings.ToLower(strings.TrimSpace(cfg.Source.Mode)) {
	case "", "realtime":
		src = source.NewTSNetRealtimeSource(ts, source.RealtimeConfig{
			Logger:       sentinelLogger,
			ReconnectMin: cfg.PollBackoffMin,
			ReconnectMax: cfg.PollBackoffMax,
		})
	case "poll":
		src = source.NewTSNetSource(ts, source.DefaultTSNetFetch)
	default:
		return nil, fmt.Errorf("unsupported source.mode %q", cfg.Source.Mode)
	}
	enrollment := onboarding.NewManager(onboarding.Config{
		Mode:                     cfg.TSNet.LoginMode,
		AuthKey:                  cfg.TSNet.AuthKey,
		AuthKeySource:            cfg.TSNet.AuthKeySource,
		AllowInteractiveFallback: cfg.TSNet.AllowInteractiveFallback,
		LoginTimeout:             cfg.TSNet.LoginTimeout,
	}, onboarding.NewTSNetProvider(ts), sentinelLogger)

	m := metrics.New(prometheus.NewRegistry())
	r := app.NewRunner(cfg, src, engine, policyEngine, notifier, st, m, sentinelLogger, enrollment)

	return &runtimeDeps{
		cfg:        cfg,
		runner:     r,
		renderer:   output.NewRenderer(cfg.Output.NoColor),
		source:     src,
		notifier:   notifier,
		enrollment: enrollment,
	}, nil
}

func runOnceWithTimeout(ctx context.Context, fn func(context.Context) error) error {
	cctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return fn(cctx)
}

func printLine(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
}
