package server

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

func DelayMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		delay := r.URL.Query().Get("delay")
		if delay != "" {
			delay, err := strconv.Atoi(delay)
			if err == nil {
				logrus.Debugf("Delaying response by %dms", delay)
				select {
				case <-time.After(time.Duration(delay) * time.Millisecond):
				case <-r.Context().Done():
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func ReadRequestBodyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Query().Has("read-body") {
			var buffer bytes.Buffer

			logrus.Debug("Reading request body")
			_, err := io.Copy(&buffer, r.Body)
			if err != nil {
				http.Error(w, "error reading body", 400)
				return
			}

			r.Body = io.NopCloser(&buffer)
		}

		next.ServeHTTP(w, r)
	})
}
