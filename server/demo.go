package server

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func RegisterDemo(s *mux.Router) {
	var err error
	var server DemoServer

	server.answerJSONRegister, err = ioutil.ReadFile("data/static/register.json")
	if err != nil {
		logrus.Fatal(err)
	}

	server.answerJSONSearchOK, err = ioutil.ReadFile("data/static/search.json")
	if err != nil {
		logrus.Fatal(err)
	}

	server.answerJSONSearchFail, err = ioutil.ReadFile("data/static/error_bad_request.json")
	if err != nil {
		logrus.Fatal(err)
	}

	s.HandleFunc("/register", server.RegisterHandler)
	s.HandleFunc("/login", server.RegisterHandler)
	s.HandleFunc("/search", server.SearchHandler)
}

type DemoServer struct {
	answerJSONRegister   []byte
	answerJSONSearchOK   []byte
	answerJSONSearchFail []byte
}

func (s *DemoServer) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if randMinMax(1, 100) > 95 {
		time.Sleep(time.Duration(randMinMax(250, 350)) * time.Millisecond)
	}

	w.Write(s.answerJSONRegister)
}

func (s *DemoServer) SearchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if len(r.URL.Query()) == 0 {
		w.Write(s.answerJSONSearchOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(s.answerJSONSearchFail)
	}
}

func randMinMax(min, max int) int {
	return rand.Intn(max-min) + min
}
