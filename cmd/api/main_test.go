package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anoop-dryad/bridgehead/cmd/api/app"
	"github.com/gin-gonic/gin"
	"github.com/go-jose/go-jose/v4/testutils/assert"
)

func TestPingRoute(t *testing.T) {
	engine := app.RegisterRoutes(gin.Default())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ping/", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
