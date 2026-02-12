package notify

import (
	"context"
	"encoding/json"
	"io"
	"sync"

	"github.com/jaxxstorm/sentinel/internal/logging"
)

type StdoutSink struct {
	name string
	w    io.Writer
	mu   sync.Mutex
}

func NewStdoutSink(name string, w io.Writer) *StdoutSink {
	return &StdoutSink{name: name, w: w}
}

func (s *StdoutSink) Name() string { return s.name }

func (s *StdoutSink) Send(_ context.Context, n Notification) error {
	payload, err := json.Marshal(struct {
		LogSource string `json:"log_source"`
		Sink      string `json:"sink"`
		Notification
	}{
		LogSource:    logging.LogSourceSink,
		Sink:         s.name,
		Notification: n,
	})
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err = s.w.Write(append(payload, '\n'))
	return err
}
