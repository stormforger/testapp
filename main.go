package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stormforger/testapp/internal/ulimit"
	"github.com/stormforger/testapp/server"
)

func main() {
	port := getEnv("PORT", "8080")
	portTLS := getEnv("TLS_PORT", "8443")
	shutdownCode := os.Getenv("SHUTDOWN_CODE")
	if shutdownCode == "" {
		logrus.Warn("SHUTDOWN_CODE not configured!")
	}

	httpReadTimeout, err := time.ParseDuration(getEnv("HTTP_READ_TIMEOUT", "15s"))
	if err != nil {
		logrus.WithError(err).Fatal("HTTP_READ_TIMEOUT parsing failed")
	}
	httpWriteTimeout, err := time.ParseDuration(getEnv("HTTP_WRITE_TIMEOUT", "15s"))
	if err != nil {
		logrus.WithError(err).Fatal("HTTP_WRITE_TIMEOUT parsing failed")
	}

	disableTLS := getEnv("DISABLE_TLS", "false") == "true"
	serverCertificateFile := getEnv("TLS_CERT", "data/pki/server.cert.pem")
	serverPrivateKeyFile := getEnv("TLS_KEY", "data/pki/server.key.pem")

	tlsConnectionInspection := getEnv("TLS_DEBUG", "0")

	ctx, cancel := context.WithCancel(context.Background())

	if _, _, _, err := ulimit.SetNoFileLimitToMax(); err != nil {
		logrus.WithError(err).Error("failed to change ulimit")
	}

	r := mux.NewRouter()
	// Install our command routes
	x := r.PathPrefix("/cmd").Subrouter()
	x.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if shutdownCode == "" || r.URL.Query().Get("code") != shutdownCode {
			http.Error(w, "Forbidden - code required", http.StatusForbidden)
			return
		}

		w.Write([]byte("OK"))
		cancel() // signal the shutdown workers
	})

	// Demo Server Routes
	r.Use(server.DelayMiddleware)
	r.Use(handlers.CompressHandler)
	server.RegisterTestAppRoutes(r)
	if !disableTLS {
		server.RegisterX509Routes(r, serverCertificateFile, serverPrivateKeyFile)
	}
	server.RegisterStaticHandler(r)

	// HTTP Server
	httpServer := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: httpWriteTimeout,
		ReadTimeout:  httpReadTimeout,
	}

	logrus.Infof("Starting HTTP server at :%s", port)
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logrus.Fatal(err)
		}
	}()

	if !disableTLS {
		httpsServer := &http.Server{
			Handler:      r,
			Addr:         ":" + portTLS,
			WriteTimeout: httpWriteTimeout,
			ReadTimeout:  httpReadTimeout,
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

		go func() {
			select {
			case <-ctx.Done():
				logrus.Info("Shutting down https server on request")
				httpsServer.Shutdown(context.Background())
			}
		}()
	}

	select {
	case <-ctx.Done():
		logrus.Info("Shutting down http server on request")
		httpServer.Shutdown(context.Background())
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
