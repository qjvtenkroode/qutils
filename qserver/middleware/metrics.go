package middleware

import (
	"net/http"

	"github.com/qjvtenkroode/qutils/qmetrics"
)

type MetricsMw struct {
	handler http.Handler
	metrics *qmetrics.Metrics
}

func (mmw MetricsMw) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mmw.handler.ServeHTTP(w, r)
}

func NewMetricsMw(handler http.Handler, m *qmetrics.Metrics) *MetricsMw {
	return &MetricsMw{handler, m}
}
