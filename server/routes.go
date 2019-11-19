package server

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func RegisterTestAppRoutes(r *mux.Router) {
	r.Use(DelayMiddleware)
	r.Use(handlers.CompressHandler)

	// demo router
	s := r.PathPrefix("/demo").Subrouter()
	RegisterDemo(s)

	// static data
	r.PathPrefix("/data/").Handler(http.StripPrefix("/data/", http.FileServer(http.Dir("data"))))

	// other handlers
	r.HandleFunc("/respond-with/bytes", RespondWithBytesHandler)
	r.HandleFunc("/do-not-respond", DoNotRespondHandler)

	// echo handler for everything else
	r.PathPrefix("/").HandlerFunc(EchoHandler)
}
