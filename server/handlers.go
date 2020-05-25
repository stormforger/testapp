package server

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/stormforger/testapp/internal/randutil"
)

func DoNotRespondHandler(w http.ResponseWriter, r *http.Request) {
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

func RespondWithBytesHandler(w http.ResponseWriter, r *http.Request) {
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

// SetCookieHandler serves a response with a cookie.
func SetCookieHandler(w http.ResponseWriter, req *http.Request) {
	cookieName := req.FormValue("cookie")
	if cookieName == "" {
		cookieName = "sessionid"
	}

	value := fmt.Sprintf("%d", randMinMax(1_000_000, 9_000_000))

	http.SetCookie(w, &http.Cookie{
		Name:  cookieName,
		Value: value,
	})

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Cookie served.")
}

// RequiresCookieHandler serves a forbidden response if no cookie is found. every cookie is accepted.
func RequiresCookieHandler(w http.ResponseWriter, req *http.Request) {
	cookieName := req.FormValue("cookie")
	if cookieName == "" {
		cookieName = "sessionid"
	}

	c, err := req.Cookie(cookieName)
	if err == http.ErrNoCookie {
		http.Error(w, "cookie required", http.StatusForbidden)
		return
	}

	fmt.Fprintf(w, "Hello. Your cookie is %s=%s", c.Name, c.Value)
}

// RandomTokenJson returns a JSON with a new random "token" value for each request.
func RandomTokenJson(w http.ResponseWriter, req *http.Request) {
	data := map[string]string{
		"token": randutil.StringWithCharset(12, randutil.Lowercase+randutil.Digits),
	}

	w.Header().Set("Content-Type", "application/json")
	e := json.NewEncoder(w)
	e.SetIndent("", " ")
	e.Encode(data) // ignore errors for now
}
