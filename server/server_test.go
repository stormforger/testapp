package server_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stormforger/testapp/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestAppHTTPServer(t *testing.T) {
	require.Nil(t, os.Chdir("..")) // our server package assumes it has access to the `data/` path directly.

	r := mux.NewRouter()
	server.RegisterStaticHandler(r)

	s := httptest.NewServer(r)

	resp, err := http.Get(s.URL)
	require.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
