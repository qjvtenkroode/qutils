package qserver

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"qkroode.nl/qutils/qdispatcher"
	"qkroode.nl/qutils/qdispatcher/worker"
	"qkroode.nl/qutils/qlog"
	"qkroode.nl/qutils/qmetrics"
	"qkroode.nl/qutils/qserver/middleware"
)

type Config struct {
    Addr string 
    Dispatcher bool
    TemplatesDir string
}

type Qserver struct {
	mux *http.ServeMux
    dispatcher *qdispatcher.Qdispatcher 
    metrics *qmetrics.Metrics
}

func NewServer() Qserver {
	q := Qserver{}
	qlog.Info("Setting up ServeMux")
	q.mux = http.NewServeMux()
    qlog.Info("Setting up qmetrics prometheus")
    reg := prometheus.NewRegistry()
    q.metrics = qmetrics.NewMetrics(reg)
	q.AddRoute("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	return q
}

func (q Qserver) AddRoute(endpoint string, handler http.Handler) {
	qlog.Info(fmt.Sprintf("Adding handler to endpoint: %s", endpoint))
	q.mux.Handle(endpoint, middleware.NewLoggingMw(handler))
}

func (q Qserver) StartServer(c Config) {
    if c.Dispatcher {
        qlog.Info("Setting up qdispatcher")
        dispatchch := make(chan worker.Message)
        q.dispatcher, _ = qdispatcher.NewQdispatcher(dispatchch, q.metrics)
        q.AddRoute("/message/", q.dispatcher.MessageHandler())
        defer q.dispatcher.Stop()
        go q.dispatcher.Start()
    }
	qlog.Info("Reading templates")

	qlog.Info("Starting webserver")
	err := http.ListenAndServe(c.Addr, q.mux)
	if errors.Is(err, http.ErrServerClosed) {
		qlog.Info(fmt.Sprint(err))
	} else if err != nil {
		qlog.Fatal(fmt.Sprint(err))
	}
}
