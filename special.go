package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
)

func respondWithBytesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")

	sizeParam := r.URL.Query().Get("size")
	size := 0
	if sizeParam != "" {
		var err error
		size, err = strconv.Atoi(sizeParam)
		if err != nil {
			size = 0
		}
	}

	w.WriteHeader(http.StatusOK)

	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		fmt.Fprintf(w, "Could not generate random response payload")
	}
	w.Write(data)
}

func doNotRespondHandler(w http.ResponseWriter, r *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}
	conn, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer conn.Close()
}
