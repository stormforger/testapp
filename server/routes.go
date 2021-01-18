package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func RegisterX509Routes(r *mux.Router, serverCertificateFile, serverPrivateKeyFile string) {
	// X.509 and EST routes
	// --------------------------------------------------------------------------
	if serverCertificateFile != "" && serverPrivateKeyFile != "" {
		err := RegisterX509ESTHandlers(r, serverCertificateFile, serverPrivateKeyFile)
		if err != nil {
			logrus.Fatal(err)
		}
	} else {
		logrus.Warn("RegisterX509Routes: empty tls certificate")
	}
}

func RegisterTestAppRoutes(r *mux.Router) {
	s := r.PathPrefix("/demo").Subrouter()
	RegisterDemo(s)

	r.PathPrefix("/data/").Handler(http.StripPrefix("/data/", http.FileServer(http.Dir("data/static"))))

	r.Path("/cookie/set").HandlerFunc(SetCookieHandler)
	r.Path("/cookie/get").HandlerFunc(RequiresCookieHandler)
}

// RegisterStaticHandler adds mostly deterministic handlers that do not rely on state or local files.
func RegisterStaticHandler(r *mux.Router) {
	r.HandleFunc("/random/get_token", RandomTokenJSON)
	r.HandleFunc("/respond-with/bytes", RespondWithBytesHandler)
	r.HandleFunc("/do-not-respond", DoNotRespondHandler)
	r.HandleFunc("/x509/inspect", clientCertInspectHandler)

	// echo handler for everything else
	r.PathPrefix("/").HandlerFunc(EchoHandler)
}
