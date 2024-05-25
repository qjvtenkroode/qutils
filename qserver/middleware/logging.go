package middleware

import (
	"net/http"

	"github.com/qjvtenkroode/qutils/qlog"
)

type LoggingMw struct {
	handler http.Handler
}

func (lmw LoggingMw) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	qlog.AccessLog(r)
	lmw.handler.ServeHTTP(w, r)
}

func NewLoggingMw(handler http.Handler) *LoggingMw {
	return &LoggingMw{handler}
}
