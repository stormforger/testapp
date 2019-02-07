package main

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	port := "8080"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	portTLS := "8443"
	if os.Getenv("TLS_PORT") != "" {
		portTLS = os.Getenv("TLS_PORT")
	}

	shutdownCode := os.Getenv("SHUTDOWN_CODE")
	if shutdownCode == "" {
		logrus.Warn("SHUTDOWN_CODE not configured!")
	}

	serverCertificateFile := os.Getenv("TLS_CERT")
	if serverCertificateFile == "" {
		serverCertificateFile = "data/pki/server.cert.pem"
	}

	serverPrivateKeyFile := os.Getenv("TLS_KEY")
	if serverPrivateKeyFile == "" {
		serverPrivateKeyFile = "data/pki/server.key.pem"
	}

	shutdownCh := make(chan bool)
	r := mux.NewRouter()

	// global middlewares
	// --------------------------------------------------------------------------
	r.Use(delayMiddleware)

	// demo router
	// --------------------------------------------------------------------------
	s := r.PathPrefix("/demo").Subrouter()
	s.HandleFunc("/register", registerHandler)
	s.HandleFunc("/search", searchHandler)

	// command router
	// --------------------------------------------------------------------------
	x := r.PathPrefix("/cmd").Subrouter()
	x.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if shutdownCode != "" && r.URL.Query().Get("code") == shutdownCode {
			w.Write([]byte("OK"))
			shutdownCh <- true
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	})

	// X.509 and EST routes
	// --------------------------------------------------------------------------
	caCertPEMData, err := ioutil.ReadFile(serverCertificateFile)
	if err != nil {
		logrus.Fatal(err)
	}
	caPrivateKeyPEMData, err := ioutil.ReadFile(serverPrivateKeyFile)
	if err != nil {
		logrus.Fatal(err)
	}
	err = configureX509Handlers(r, caCertPEMData, caPrivateKeyPEMData)
	if err != nil {
		logrus.Fatal(err)
	}

	// static data
	// --------------------------------------------------------------------------
	r.PathPrefix("/data/").Handler(http.StripPrefix("/data/", http.FileServer(http.Dir("data/static"))))

	// other handlers
	// --------------------------------------------------------------------------
	r.HandleFunc("/respond-with/bytes", respondWithBytesHandler)
	r.HandleFunc("/do-not-respond", doNotRespondHandler)

	// echo handler for everything else
	// --------------------------------------------------------------------------
	r.PathPrefix("/").HandlerFunc(echoHandler)

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

	logrus.Infof("Starting HTTPS server at :%s", portTLS)
	go func() {
		err := httpsServer.ListenAndServeTLS(serverCertificateFile, serverPrivateKeyFile)
		if err != nil && err != http.ErrServerClosed {
			logrus.Fatal(err)
		}
	}()

	select {
	case <-shutdownCh:
		httpServer.Shutdown(context.Background())
		logrus.Info("Shutting down on request")
	}
}
