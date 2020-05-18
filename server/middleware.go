package server

import (
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
