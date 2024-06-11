package qdispatcher

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/qjvtenkroode/qutils/qdispatcher/worker"
	"github.com/qjvtenkroode/qutils/qlog"
	"github.com/qjvtenkroode/qutils/qmetrics"
)

type Qdispatcher struct {
	/*
	   Qdispatcher will accept Messages and dispatch those to a registered Worker.
	   Workers have a queue that holds messages to process.
	*/
	workers    map[string]worker.Worker
	dispatchch chan worker.Message
	quitch     chan struct{}
	metrics    *qmetrics.Metrics
}

func NewQdispatcher(dispatchch chan worker.Message, metrics *qmetrics.Metrics) (*Qdispatcher, error) {
	chat_id := os.Getenv("hivemind_chat_id")
    if chat_id == "" {
        qlog.Error("Environment variable hivemind_chat_id not set or empty")
    }
	access_token := os.Getenv("hivemind_access_token")
    if access_token == "" {
        qlog.Error("Environment variable hivemind_access_token not set or empty")
    }
	workers := make(map[string]worker.Worker)
	workers["telegram"] = worker.NewTelegramWorker(metrics, chat_id, access_token)
	workers["metric"] = worker.NewMetricStoreWorker(metrics)
	return &Qdispatcher{
		workers:    workers,
		dispatchch: dispatchch,
		quitch:     make(chan struct{}),
		metrics:    metrics,
	}, nil
}

func (q Qdispatcher) MessageHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		topic := strings.Split(r.URL.Path, "/")[2]
		payload := ""
		if r.Body != nil {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			payload = string(body)
		}
		m := worker.Message{Topic: topic, Payload: payload}
		q.AddMessage(m)

		w.WriteHeader(http.StatusOK)
	}

	return http.HandlerFunc(fn)
}

func (q Qdispatcher) AddMessage(m worker.Message) {
	q.dispatchch <- m
	q.metrics.Messages.With(prometheus.Labels{"worker": "dispatcher"}).Inc()
}

func (q Qdispatcher) Start() {
	qlog.Info("Starting Qdispatcher")
	for k, v := range q.workers {
		qlog.Info(fmt.Sprintf("Initializing topic: %s", k))
		go v.Start()
	}
	q.loop()
}

func (q Qdispatcher) Stop() {
	for k, v := range q.workers {
		qlog.Info(fmt.Sprintf("Initialize shutdown of topic: %s", k))
		v.Stop()
	}
	q.quitch <- struct{}{}
}

func (q Qdispatcher) loop() {
	for {
		select {
		case <-q.quitch:
			qlog.Info("Stopping Qdispatcher")
			return
		case msg := <-q.dispatchch:
			qlog.Info(fmt.Sprintf("Got message to dispatch: %s", msg))
			if w, ok := q.workers[msg.Topic]; ok {
				w.AddMessage(msg)
			} else {
				qlog.Error(fmt.Sprintf("Topic doesn't exist: %s", msg))
                q.metrics.Errors.With(prometheus.Labels{"worker": "dispatcher"}).Inc()
			}
		}
	}
}
