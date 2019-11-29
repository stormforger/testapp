package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stormforger/testapp/server"
)

func main() {
	port := getEnv("PORT", "8080")
	portTLS := getEnv("TLS_PORT", "8443")
	shutdownCode := os.Getenv("SHUTDOWN_CODE")
	if shutdownCode == "" {
		logrus.Warn("SHUTDOWN_CODE not configured!")
	}

	serverCertificateFile := getEnv("TLS_CERT", "data/pki/server.cert.pem")
	serverPrivateKeyFile := getEnv("TLS_KEY", "data/pki/server.key.pem")

	tlsConnectionInspection := getEnv("TLS_DEBUG", "0")

	shutdownCh := make(chan bool)

	r := mux.NewRouter()
	// Install our command routes
	x := r.PathPrefix("/cmd").Subrouter()
	x.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if shutdownCode != "" && r.URL.Query().Get("code") == shutdownCode {
			w.Write([]byte("OK"))
			shutdownCh <- true
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	})

	// Demo Server Routes
	server.RegisterTestAppRoutes(r, serverCertificateFile, serverPrivateKeyFile)

	httpServer := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logrus.Infof("Starting HTTP server at :%s", port)
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logrus.Fatal(err)
		}
	}()

	httpsServer := &http.Server{
		Handler:      r,
		Addr:         ":" + portTLS,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			ClientAuth:         tls.RequestClientCert,
		},
	}

	if tlsConnectionInspection == "1" {
		setupTLSConnectionInspection(httpsServer)
	}

	logrus.Infof("Starting HTTPS server at :%s", portTLS)
	go func() {
		err := httpsServer.ListenAndServeTLS(serverCertificateFile, serverPrivateKeyFile)
		if err != nil && err != http.ErrServerClosed {
			logrus.Fatal(err)
		}
	}()

	select {
	case <-shutdownCh:
		logrus.Info("Shutting down on request")
		httpServer.Shutdown(context.Background())
		httpsServer.Shutdown(context.Background())
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
