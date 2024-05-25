package qmetrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Messages *prometheus.CounterVec
	Errors *prometheus.CounterVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		Messages: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "qdispatcher",
			Name:      "messages",
			Help:      "Number of messages processed.",
		}, []string{"worker"}),
        Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "qdispatcher",
			Name:      "errors",
			Help:      "Number of messages failed processing.",
		}, []string{"worker"}),

	}
	reg.MustRegister(m.Messages, m.Errors)
	return m
}
