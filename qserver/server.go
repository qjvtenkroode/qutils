package qserver

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/qjvtenkroode/qutils/qlog"
	"github.com/qjvtenkroode/qutils/qmetrics"
	"github.com/qjvtenkroode/qutils/qserver/middleware"
)

type Config struct {
	Addr         string
	TemplatesDir string
}

type Qserver struct {
	mux     *http.ServeMux
	metrics *qmetrics.Metrics
}

func NewServer() Qserver {
	q := Qserver{}
	qlog.Info("Setting up ServeMux")
	q.mux = http.NewServeMux()
	qlog.Info("Setting up qmetrics prometheus")
	reg := prometheus.NewRegistry()
	q.metrics = qmetrics.NewMetrics(reg)
	q.AddRoute("GET /metrics/", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	return q
}

func (q Qserver) AddRoute(endpoint string, handler http.Handler) {
	qlog.Info(fmt.Sprintf("Adding handler to endpoint: %s", endpoint))
	q.mux.Handle(endpoint, middleware.NewLoggingMw(handler))
}

func (q Qserver) StartServer(c Config) {
	qlog.Info("Reading templates")

	qlog.Info("Starting webserver")
	err := http.ListenAndServe(c.Addr, q.mux)
	if errors.Is(err, http.ErrServerClosed) {
		qlog.Info(fmt.Sprint(err))
	} else if err != nil {
		qlog.Fatal(fmt.Sprint(err))
	}
}
