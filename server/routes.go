package server

import (
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
	if serverCertificateFile != "" && serverPrivateKeyFile != "" {
		err := RegisterX509ESTHandlers(r, serverCertificateFile, serverPrivateKeyFile)
		if err != nil {
			logrus.Fatal(err)
		}
	} else {
		logrus.Warn("RegisterTestAppRoutes: empty tls certificate")
	}

	s := r.PathPrefix("/demo").Subrouter()
	RegisterDemo(s)

	r.PathPrefix("/data/").Handler(http.StripPrefix("/data/", http.FileServer(http.Dir("data/static"))))

	r.Path("/cookie/set").HandlerFunc(SetCookieHandler)
	r.Path("/cookie/get").HandlerFunc(RequiresCookieHandler)
	RegisterStaticHandler(r)
}

func RegisterStaticHandler(r *mux.Router) {
	r.HandleFunc("/respond-with/bytes", RespondWithBytesHandler)
	r.HandleFunc("/do-not-respond", DoNotRespondHandler)

	// echo handler for everything else
	r.PathPrefix("/").HandlerFunc(EchoHandler)
}
