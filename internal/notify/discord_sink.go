package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	discordEmbedTitleLimit       = 256
	discordEmbedDescriptionLimit = 4096
	discordEmbedFieldLimit       = 1024
	discordPayloadSummaryTrim    = 860
)

type DiscordSink struct {
	name       string
	url        string
	client     *http.Client
	maxRetries int
	backoff    time.Duration
	logger     *zap.Logger
}

type discordPayload struct {
	Embeds []discordEmbed `json:"embeds"`
}

type discordEmbed struct {
	Title       string             `json:"title,omitempty"`
	URL         string             `json:"url,omitempty"`
	Description string             `json:"description,omitempty"`
	Color       int                `json:"color,omitempty"`
	Timestamp   string             `json:"timestamp,omitempty"`
	Fields      []discordEmbedItem `json:"fields,omitempty"`
}

type discordEmbedItem struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

func NewDiscordSink(name, url string, logger *zap.Logger) *DiscordSink {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &DiscordSink{
		name:       name,
		url:        url,
		client:     &http.Client{Timeout: 10 * time.Second},
		maxRetries: 3,
		backoff:    200 * time.Millisecond,
		logger:     logger,
	}
}

func (s *DiscordSink) Name() string { return s.name }

func (s *DiscordSink) Send(ctx context.Context, n Notification) error {
	payload, err := json.Marshal(discordWebhookPayload(n))
	if err != nil {
		return err
	}

	var lastErr error
	for i := 0; i <= s.maxRetries; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", n.IdempotencyKey)

		resp, err := s.client.Do(req)
		if err == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			s.logger.Info("discord send succeeded",
				zap.String("sink", s.name),
				zap.Int("status_code", resp.StatusCode),
			)
			_ = resp.Body.Close()
			return nil
		}
		if resp != nil {
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("unexpected status %d", resp.StatusCode)
			s.logger.Warn("discord send failed",
				zap.String("sink", s.name),
				zap.Int("status_code", resp.StatusCode),
				zap.Int("attempt", i+1),
				zap.Int("max_attempts", s.maxRetries+1),
			)
		} else {
			lastErr = err
			s.logger.Warn("discord send failed",
				zap.String("sink", s.name),
				zap.Error(err),
				zap.Int("attempt", i+1),
				zap.Int("max_attempts", s.maxRetries+1),
			)
		}
		if i < s.maxRetries {
			select {
			case <-ctx.Done():
				s.logger.Warn("discord send canceled",
					zap.String("sink", s.name),
					zap.Int("attempt", i+1),
					zap.Int("max_attempts", s.maxRetries+1),
				)
				return ctx.Err()
			case <-time.After(s.backoff * time.Duration(i+1)):
			}
		}
	}
	return fmt.Errorf("discord sink failed after retries: %w", lastErr)
}

func discordWebhookPayload(n Notification) discordPayload {
	return discordPayload{
		Embeds: []discordEmbed{discordEmbedForEvent(n)},
	}
}

func discordEmbedForEvent(n Notification) discordEmbed {
	evt := n.Event
	title := truncateString("Sentinel "+evt.EventType, discordEmbedTitleLimit)
	desc := truncateString(
		fmt.Sprintf("**Subject** `%s/%s`\n**Severity** `%s`", evt.SubjectType, evt.SubjectID, evt.Severity),
		discordEmbedDescriptionLimit,
	)

	return discordEmbed{
		Title:       title,
		URL:         "https://login.tailscale.com/admin/machines",
		Description: desc,
		Color:       discordSeverityColor(evt.Severity),
		Timestamp:   evt.Timestamp.UTC().Format(time.RFC3339Nano),
		Fields: []discordEmbedItem{
			{
				Name:   "Event Type",
				Value:  fmt.Sprintf("`%s`", evt.EventType),
				Inline: true,
			},
			{
				Name:   "Subject",
				Value:  fmt.Sprintf("`%s/%s`", evt.SubjectType, evt.SubjectID),
				Inline: true,
			},
			{
				Name:  "Payload",
				Value: discordPayloadFieldValue(evt.Payload),
			},
		},
	}
}

func discordPayloadFieldValue(payload map[string]any) string {
	value := "{}"
	if len(payload) > 0 {
		if b, err := json.MarshalIndent(payload, "", "  "); err == nil {
			value = string(b)
		}
	}
	value = strings.TrimSpace(value)
	if len(value) > discordPayloadSummaryTrim {
		value = value[:discordPayloadSummaryTrim] + "..."
	}
	wrapped := "```json\n" + value + "\n```"
	return truncateString(wrapped, discordEmbedFieldLimit)
}

func truncateString(in string, limit int) string {
	if limit <= 0 || len(in) <= limit {
		return in
	}
	if limit <= 3 {
		return in[:limit]
	}
	return in[:limit-3] + "..."
}

func discordSeverityColor(sev string) int {
	switch strings.ToLower(strings.TrimSpace(sev)) {
	case "error":
		return 0xE74C3C
	case "warn", "warning":
		return 0xF39C12
	default:
		return 0x3498DB
	}
}
