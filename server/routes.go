package server

import (
	"io/ioutil"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func RegisterTestAppRoutes(r *mux.Router, serverCertificateFile, serverPrivateKeyFile string) {
	r.Use(DelayMiddleware)
	r.Use(handlers.CompressHandler)

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
	err = RegisterX509Handlers(r, caCertPEMData, caPrivateKeyPEMData)
	if err != nil {
		logrus.Fatal(err)
	}

	// demo router
	// --------------------------------------------------------------------------
	s := r.PathPrefix("/demo").Subrouter()
	RegisterDemo(s)

	// static data
	// --------------------------------------------------------------------------
	r.PathPrefix("/data/").Handler(http.StripPrefix("/data/", http.FileServer(http.Dir("data/static"))))

	// other handlers
	// --------------------------------------------------------------------------
	r.HandleFunc("/respond-with/bytes", RespondWithBytesHandler)
	r.HandleFunc("/do-not-respond", DoNotRespondHandler)

	// echo handler for everything else
	r.PathPrefix("/").HandlerFunc(EchoHandler)
}
