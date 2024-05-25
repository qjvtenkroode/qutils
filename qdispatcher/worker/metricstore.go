package worker

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/prometheus/client_golang/prometheus"
	"qkroode.nl/qutils/qlog"
	"qkroode.nl/qutils/qmetrics"
)

var dbname = "test.bdb"
var dbperms os.FileMode = 0770
var options = &bolt.Options{Timeout: 1 * time.Second}

type Metric struct {
	// Id field is a millisecond timestamp,
	// Name field corresponds to the metric name,
	// Instance field is the entity generating the metric.
	Id       time.Time     `json:"id"`
	Value    float64 `json:"value"`
	Name     string  `json:"name"`
	Instance string  `json:"instance"`
}

type MetricStoreWorker struct {
	messagech chan Message
	quitch    chan struct{}
    metrics *qmetrics.Metrics
    database *bolt.DB 
}

func NewMetricStoreWorker(metrics *qmetrics.Metrics) *MetricStoreWorker {
	messagech := make(chan Message)
	return &MetricStoreWorker{
		messagech: messagech,
		quitch:    make(chan struct{}),
        metrics: metrics,
        database: nil,
	}
}

func (ms MetricStoreWorker) AddMessage(m Message) {
	// Expects a JSON payload which corresponds to type Metric
	ms.messagech <- m
    ms.metrics.Messages.With(prometheus.Labels{"worker": "metricstore"}).Inc()
}

func (ms MetricStoreWorker) Start() {
    var err error
    ms.database, err = bolt.Open(dbname, dbperms, options)
    if err != nil {
        qlog.Fatal(fmt.Sprintf("MetricStoreWorker - could not open boltdb - %v", err))
    }
    defer ms.database.Close()
	qlog.Info("Starting MetricStoreWorker")
	ms.loop()
}

func (ms MetricStoreWorker) Stop() {
	ms.quitch <- struct{}{}
}

func (ms MetricStoreWorker) processMessage(m Message) error {
	// Processes a Message into a Metric for storage
	metric := Metric{}
    err := json.Unmarshal([]byte(m.Payload), &metric)
    if err != nil {
        qlog.Error(fmt.Sprintf("MetricStoreWorker: could not unmarshal json to metric struct - %v", err))
        return err
    }
	err = ms.store(metric)
	return err
}

func (ms MetricStoreWorker) store(me Metric) error {
    // Marshal the struct into json for storage
    metricBytes, err := json.Marshal(me)
    if err != nil {
        qlog.Error(fmt.Sprintf("MetricStoreWorker: could not marshal metric json - %v", err))
        return err
    }
	// Stores a metric in a BoltDB database
    err = ms.database.Update(func(tx *bolt.Tx) error {
        b, err := tx.CreateBucketIfNotExists([]byte(me.Name))
        if err != nil {
            qlog.Error(fmt.Sprintf("MetricStoreWorker: could not create bucket - %v", err))
            return err
        }
        bb, err := b.CreateBucketIfNotExists([]byte(me.Instance))
        if err != nil {
            qlog.Error(fmt.Sprintf("MetricStoreWorker: could not create bucket - %v", err))
            return err
        }
        err = bb.Put([]byte(me.Id.Format(time.RFC3339Nano)), metricBytes)
        return err
    })
	return err
}

func (ms MetricStoreWorker) loop() {
	for {
		select {
		case <-ms.quitch:
			qlog.Info("Stopping MetricStoreWorker")
			return
		case msg := <-ms.messagech:
			qlog.Info(fmt.Sprintf("MetricStoreWorker: got message: %s", msg))
            err := ms.processMessage(msg)
            if err != nil {
                ms.metrics.Errors.With(prometheus.Labels{"worker": "metricstore"}).Inc()
            }
		}
	}
}
