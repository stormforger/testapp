package main

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

var answerJSONRegister []byte
var answerJSONSearchOK []byte
var answerJSONSearchFail []byte

func registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if randMinMax(1, 100) > 95 {
		time.Sleep(time.Duration(randMinMax(250, 350)) * time.Millisecond)
	}

	w.Write(answerJSONRegister)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if len(r.URL.Query()) == 0 {
		w.Write(answerJSONSearchOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(answerJSONSearchFail)
	}
}

func init() {
	var err error

	answerJSONRegister, err = ioutil.ReadFile("data/register.json")
	if err != nil {
		logrus.Fatal(err)
	}

	answerJSONSearchOK, err = ioutil.ReadFile("data/search.json")
	if err != nil {
		logrus.Fatal(err)
	}

	answerJSONSearchFail, err = ioutil.ReadFile("data/error_bad_request.json")
	if err != nil {
		logrus.Fatal(err)
	}
}

func randMinMax(min, max int) int {
	return rand.Intn(max-min) + min
}
