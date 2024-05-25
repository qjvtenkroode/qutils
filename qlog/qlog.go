package qlog

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func Debug(msg string) {
	log.Printf("- DEBUG - %s", msg)
}

func Info(msg string) {
	log.Printf("- INFO - %s", msg)
}

func Warn(msg string) {
	log.Printf("- WARN - %s", msg)
}

func Error(msg string) {
	log.Printf("- ERROR - %s", msg)
}

func Fatal(msg string) {
	log.Fatalf("- FATAL - %s", msg)
	os.Exit(1)
}

func AccessLog(r *http.Request) {
	Info(fmt.Sprintf("method=%s uri=%s", r.Method, r.URL.Path))
}
