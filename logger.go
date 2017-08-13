package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/op/go-logging"
)

func initializeLogger() *logging.Logger {

	var format = logging.MustStringFormatter(
		fmt.Sprintf(
			"%v %v",
			"%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.5s}",
			"%{id:03x}%{color:reset} %{message}",
		),
	)

	var backend = logging.NewBackendFormatter(
		logging.NewLogBackend(os.Stdout, "", 0),
		format)

	logging.SetBackend(backend)

	return logging.MustGetLogger("logger")

}

func DecorateWithLogger(inner http.Handler, name string) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Debugf(
			"%s\t%s (▶ %s)",
			r.Method,
			r.RequestURI,
			name,
		)

		inner.ServeHTTP(w, r)

		log.Debugf(
			"%s\t%s (▶ %s) (%s)",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})

}
