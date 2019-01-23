package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
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
	r.Use(delayMiddleware)

	// demo router
	s := r.PathPrefix("/demo").Subrouter()
	s.HandleFunc("/register", registerHandler)
	s.HandleFunc("/search", searchHandler)

	// command router
	x := r.PathPrefix("/cmd").Subrouter()
	x.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if shutdownCode != "" && r.URL.Query().Get("code") == shutdownCode {
			w.Write([]byte("OK"))
			shutdownCh <- true
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	})

	// static data
	r.PathPrefix("/data/").Handler(http.StripPrefix("/data/", http.FileServer(http.Dir("data"))))

	// other handlers
	r.HandleFunc("/respond-with/bytes", respondWithBytesHandler)
	r.HandleFunc("/do-not-respond", doNotRespondHandler)

	// echo handler for everything else
	r.PathPrefix("/").HandlerFunc(echoHandler)

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

// FIXME: this will print a stack trace to the logs, but I don't know
//        if there is another way to kill the request (and the connection)
//        in a way that we do not respond
func doNotRespondHandler(w http.ResponseWriter, r *http.Request) {
	panic("do-not-respond")
}

func respondWithBytesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")

	sizeParam := r.URL.Query().Get("size")
	size := 0
	if sizeParam != "" {
		var err error
		size, err = strconv.Atoi(sizeParam)
		if err != nil {
			size = 0
		}
	}

	w.WriteHeader(200)

	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		fmt.Fprintf(w, "Could not generate random response payload")
	}
	w.Write(data)
}
