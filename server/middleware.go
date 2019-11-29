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
				logrus.Infof("Delaying response by %dms", delay)
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}
		}

		next.ServeHTTP(w, r)
	})
}
