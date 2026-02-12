package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	NetmapPollsTotal          prometheus.Counter
	NetmapPollDurationSeconds prometheus.Histogram
	DiffsDetectedTotal        *prometheus.CounterVec
	EventsEmittedTotal        *prometheus.CounterVec
	NotificationsSentTotal    *prometheus.CounterVec
	NotificationsSuppressed   *prometheus.CounterVec
	StateStoreErrorsTotal     prometheus.Counter
}

func New(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		NetmapPollsTotal:          prometheus.NewCounter(prometheus.CounterOpts{Name: "netmap_polls_total", Help: "Total netmap polls"}),
		NetmapPollDurationSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{Name: "netmap_poll_duration_seconds", Help: "Netmap poll duration"}),
		DiffsDetectedTotal:        prometheus.NewCounterVec(prometheus.CounterOpts{Name: "diffs_detected_total", Help: "Diffs detected by type"}, []string{"type"}),
		EventsEmittedTotal:        prometheus.NewCounterVec(prometheus.CounterOpts{Name: "events_emitted_total", Help: "Events emitted by type"}, []string{"type"}),
		NotificationsSentTotal:    prometheus.NewCounterVec(prometheus.CounterOpts{Name: "notifications_sent_total", Help: "Notifications sent by sink"}, []string{"sink"}),
		NotificationsSuppressed:   prometheus.NewCounterVec(prometheus.CounterOpts{Name: "notifications_suppressed_total", Help: "Suppressed notifications by reason"}, []string{"reason"}),
		StateStoreErrorsTotal:     prometheus.NewCounter(prometheus.CounterOpts{Name: "state_store_errors_total", Help: "State store errors"}),
	}
	reg.MustRegister(
		m.NetmapPollsTotal,
		m.NetmapPollDurationSeconds,
		m.DiffsDetectedTotal,
		m.EventsEmittedTotal,
		m.NotificationsSentTotal,
		m.NotificationsSuppressed,
		m.StateStoreErrorsTotal,
	)
	return m
}
