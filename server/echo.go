package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
)

// EchoHandler is a simple http.Handler for debugging webrequest.
// Each request is sent back to the client as the payload.
func EchoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	// Feature: inject redirect location header
	location := r.URL.Query().Get("location")
	if location != "" {
		w.Header().Add("location", location)
	}

	// Feature: Change the response status
	answerStatus := r.URL.Query().Get("status")
	if answerStatus != "" {
		code, err := strconv.Atoi(answerStatus)
		if err != nil {
			code = 200
		}

		if code >= 100 && code <= 999 {
			w.WriteHeader(code)
		}
	}

	// Exclude request body if too large
	if r.ContentLength > 10000 {
		reqDump, err := httputil.DumpRequest(r, false)
		if err != nil {
			fmt.Fprintf(w, "Could not dump request")
			return
		}

		w.Write(reqDump)
		return
	}

	reqDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Fprintf(w, "Could not dump request")
		return
	}

	w.Write(reqDump)
}
