package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func New() *prometheus.CounterVec {
	return promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "leaderboard",
			Subsystem: "ingest",
			Name:      "events_processed_total",
			Help:      "Total number of processed events",
		},
		[]string{"result"})
}
