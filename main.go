package main

import (
	"log"
	"net/http"
	"os"
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

	r := mux.NewRouter()
	r.Use(delayMiddleware)

	s := r.PathPrefix("/demo").Subrouter()
	s.HandleFunc("/register", registerHandler)
	s.HandleFunc("/search", searchHandler)

	r.PathPrefix("/data/").Handler(http.StripPrefix("/data/", http.FileServer(http.Dir("data"))))

	r.PathPrefix("/").HandlerFunc(echoHandler)

	listenAddr := ":" + port

	srv := &http.Server{
		Handler:      r,
		Addr:         listenAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logrus.Infof("Starting at :%s", port)

	log.Fatal(srv.ListenAndServe())
}
