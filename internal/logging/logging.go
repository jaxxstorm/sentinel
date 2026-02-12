package logging

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Format  string
	Level   string
	NoColor bool
}

const (
	LogSourceField     = "log_source"
	LogSourceSentinel  = "sentinel"
	LogSourceTailscale = "tailscale"
	LogSourceSink      = "sink"
)

func NewLogger(cfg Config) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(strings.ToLower(cfg.Level))); err != nil {
		level = zapcore.InfoLevel
	}

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "timestamp"
	encCfg.MessageKey = "message"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var enc zapcore.Encoder
	if strings.EqualFold(cfg.Format, "pretty") || cfg.Format == "" {
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		if cfg.NoColor || os.Getenv("NO_COLOR") != "" {
			encCfg.EncodeLevel = zapcore.CapitalLevelEncoder
		}
		enc = zapcore.NewConsoleEncoder(encCfg)
	} else {
		encCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
		enc = zapcore.NewJSONEncoder(encCfg)
	}

	core := zapcore.NewCore(enc, zapcore.AddSync(os.Stdout), level)
	return zap.New(core), nil
}

func WithSource(logger *zap.Logger, source string) *zap.Logger {
	if logger == nil {
		logger = zap.NewNop()
	}
	return logger.With(zap.String(LogSourceField, source))
}

func LogfAdapter(logger *zap.Logger, level zapcore.Level) func(format string, args ...any) {
	l := logger
	if l == nil {
		l = zap.NewNop()
	}
	return func(format string, args ...any) {
		msg := strings.TrimSpace(fmt.Sprintf(format, args...))
		if msg == "" {
			return
		}
		switch level {
		case zapcore.DebugLevel:
			l.Debug(msg)
		case zapcore.InfoLevel:
			l.Info(msg)
		case zapcore.WarnLevel:
			l.Warn(msg)
		case zapcore.ErrorLevel:
			l.Error(msg)
		default:
			l.Info(msg)
		}
	}
}

func RedactPeerMetadata(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		switch strings.ToLower(k) {
		case "endpoint", "ip", "user", "email", "displayname":
			out[k] = "[redacted]"
		default:
			out[k] = v
		}
	}
	return out
}

func RedactSecret(value string) string {
	v := strings.TrimSpace(value)
	if v == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(v))
	return "[redacted:" + hex.EncodeToString(sum[:4]) + "]"
}

func RedactAuthKey(value string) string {
	return RedactSecret(value)
}
