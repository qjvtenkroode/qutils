package worker

//TODO: clean up variables into config or flags

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"qkroode.nl/qutils/qlog"
	"qkroode.nl/qutils/qmetrics"
)

type TelegramWorker struct {
	messagech chan Message
	quitch    chan struct{}
    metrics *qmetrics.Metrics
    chat_id  string
    access_token string
}

func NewTelegramWorker(metrics *qmetrics.Metrics, chat_id string, access_token string) *TelegramWorker {
	messagech := make(chan Message)
	return &TelegramWorker{
		messagech: messagech,
		quitch:    make(chan struct{}),
        metrics: metrics,
        chat_id: chat_id,
        access_token: access_token,
	}
}
func (t TelegramWorker) AddMessage(m Message) {
	t.messagech <- m
    t.metrics.Messages.With(prometheus.Labels{"worker": "telegram"}).Inc()
}

func (t TelegramWorker) Start() {
	qlog.Info("Starting TelegramWorker")
	t.loop()
}

func (t TelegramWorker) Stop() {
	t.quitch <- struct{}{}
}

func (t TelegramWorker) loop() {
	for {
		select {
		case <-t.quitch:
			qlog.Info("Stopping TelegramWorker")
			return
		case msg := <-t.messagech:
            resp, err := http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s", t.access_token, t.chat_id, msg.Payload))
			qlog.Info(fmt.Sprintf("Telegram API POST response: %s for message: %s", resp.Status, msg))
			if err != nil {
				qlog.Error(fmt.Sprintf("Telegram API POST failed: %s", err))
                t.metrics.Errors.With(prometheus.Labels{"worker": "telegram"}).Inc()
			}
		}
	}
}
