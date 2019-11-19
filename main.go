package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stormforger/testapp/server"
)

func main() {
	var port string
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	} else {
		port = "9000"
	}

	shutdownCode := os.Getenv("SHUTDOWN_CODE")
	if shutdownCode == "" {
		logrus.Warn("SHUTDOWN_CODE not configured!")
	}

	shutdownCh := make(chan bool)

	r := mux.NewRouter()
	server.RegisterTestAppRoutes(r)

	// Also install our command routes
	x := r.PathPrefix("/cmd").Subrouter()
	x.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if shutdownCode != "" && r.URL.Query().Get("code") == shutdownCode {
			w.Write([]byte("OK"))
			shutdownCh <- true
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	})

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logrus.Infof("Starting at :%s", port)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	select {
	case <-shutdownCh:
		srv.Shutdown(context.Background())
		logrus.Info("Shutting down on request")
	}
}
