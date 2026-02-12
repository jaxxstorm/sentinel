package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type WebhookSink struct {
	name       string
	url        string
	client     *http.Client
	maxRetries int
	backoff    time.Duration
	logger     *zap.Logger
}

func NewWebhookSink(name, url string, logger *zap.Logger) *WebhookSink {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &WebhookSink{
		name:       name,
		url:        url,
		client:     &http.Client{Timeout: 10 * time.Second},
		maxRetries: 3,
		backoff:    200 * time.Millisecond,
		logger:     logger,
	}
}

func (s *WebhookSink) Name() string { return s.name }

func (s *WebhookSink) Send(ctx context.Context, n Notification) error {
	payload, err := json.Marshal(n)
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
			s.logger.Info("webhook send succeeded",
				zap.String("sink", s.name),
				zap.Int("status_code", resp.StatusCode),
			)
			_ = resp.Body.Close()
			return nil
		}
		if resp != nil {
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("unexpected status %d", resp.StatusCode)
			s.logger.Warn("webhook send failed",
				zap.String("sink", s.name),
				zap.Int("status_code", resp.StatusCode),
				zap.Int("attempt", i+1),
				zap.Int("max_attempts", s.maxRetries+1),
			)
		} else {
			lastErr = err
			s.logger.Warn("webhook send failed",
				zap.String("sink", s.name),
				zap.Error(err),
				zap.Int("attempt", i+1),
				zap.Int("max_attempts", s.maxRetries+1),
			)
		}
		if i < s.maxRetries {
			select {
			case <-ctx.Done():
				s.logger.Warn("webhook send canceled",
					zap.String("sink", s.name),
					zap.Int("attempt", i+1),
					zap.Int("max_attempts", s.maxRetries+1),
				)
				return ctx.Err()
			case <-time.After(s.backoff * time.Duration(i+1)):
			}
		}
	}
	return fmt.Errorf("webhook sink failed after retries: %w", lastErr)
}
