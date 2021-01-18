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

type testAppConfig struct {
	Port                  string
	PortTLS               string
	ShutdownCode          string
	HttpReadTimeout       time.Duration
	HttpWriteTimeout      time.Duration
	DisableTLS            bool
	ServerCertificateFile string
	ServerPrivateKeyFile  string
	DebugTLS              bool
}

func configFromENV() testAppConfig {
	port := getEnv("PORT", "8080")
	portTLS := getEnv("TLS_PORT", "8443")
	shutdownCode := os.Getenv("SHUTDOWN_CODE")

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

	tlsConnectionInspection := getEnv("TLS_DEBUG", "false") == "true"

	return testAppConfig{
		Port:                  port,
		PortTLS:               portTLS,
		ShutdownCode:          shutdownCode,
		HttpReadTimeout:       httpReadTimeout,
		HttpWriteTimeout:      httpWriteTimeout,
		DisableTLS:            disableTLS,
		ServerCertificateFile: serverCertificateFile,
		ServerPrivateKeyFile:  serverPrivateKeyFile,
		DebugTLS:              tlsConnectionInspection,
	}
}

func main() {
	config := configFromENV()
	if config.ShutdownCode == "" {
		logrus.Warn("SHUTDOWN_CODE not configured!")
	}

	if _, _, _, err := ulimit.SetNoFileLimitToMax(); err != nil {
		logrus.WithError(err).Error("failed to change ulimit")
	}

	ctx, cancel := context.WithCancel(context.Background()) // create a context for the shutdown handler to kill the servers
	r := provideServerHandler(config, cancel)

	if !config.DisableTLS {
		httpsServer := provideHttpsServer(r, config)

		if config.DebugTLS {
			setupTLSConnectionInspection(httpsServer)
		}

		logrus.Infof("Starting HTTPS server at %s", httpsServer.Addr)
		go func() {
			err := httpsServer.ListenAndServeTLS(config.ServerCertificateFile, config.ServerPrivateKeyFile)
			if err != nil && err != http.ErrServerClosed {
				logrus.Fatal(err)
			}
		}()

		go func() {
			<-ctx.Done()
			logrus.Info("Shutting down https server on request")
			httpsServer.Shutdown(context.Background())
		}()
	}

	// HTTP Server
	httpServer := provideHttpServer(r, config)

	logrus.Infof("Starting HTTP server at :%s", httpServer.Addr)
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logrus.Fatal(err)
		}
	}()

	<-ctx.Done()
	logrus.Info("Shutting down http server on request")
	httpServer.Shutdown(context.Background())
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func provideServerHandler(config testAppConfig, cancel context.CancelFunc) http.Handler {
	r := mux.NewRouter()
	// Install our command routes
	x := r.PathPrefix("/cmd").Subrouter()
	x.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if config.ShutdownCode == "" || r.URL.Query().Get("code") != config.ShutdownCode {
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
	if !config.DisableTLS {
		server.RegisterX509Routes(r, config.ServerCertificateFile, config.ServerPrivateKeyFile)
	}
	server.RegisterStaticHandler(r)
	return r
}

func provideHttpServer(handler http.Handler, config testAppConfig) *http.Server {
	return &http.Server{
		Handler:      handler,
		Addr:         ":" + config.Port,
		WriteTimeout: config.HttpWriteTimeout,
		ReadTimeout:  config.HttpReadTimeout,
	}
}

func provideHttpsServer(handler http.Handler, config testAppConfig) *http.Server {
	return &http.Server{
		Handler:      handler,
		Addr:         ":" + config.PortTLS,
		WriteTimeout: config.HttpWriteTimeout,
		ReadTimeout:  config.HttpReadTimeout,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			ClientAuth:         tls.RequestClientCert,
		},
	}
}
