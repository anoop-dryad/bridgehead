package server

import (
	"net/http"
	"time"

	"github.com/anoop-dryad/bridgehead/config"
	"github.com/anoop-dryad/bridgehead/infra/http/handlers"
	"github.com/anoop-dryad/bridgehead/infra/http/routes"
	"github.com/gin-gonic/gin"
)

func NewServer(cfg config.HTTP, deps handlers.Dependencies) *http.Server {
	engine := gin.New()
	engine.Use(gin.Recovery())
	routes.Register(engine, deps, cfg)
	return &http.Server{
		Addr:           ":8080",
		Handler:        engine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

}
