package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-jose/go-jose/v4/testutils/assert"
)

func TestPingRoute(t *testing.T) {
	router := getPingPong(setupRouter())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
