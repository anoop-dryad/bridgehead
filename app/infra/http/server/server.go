package server

import (
	"context"
	"net/http"
	"time"

	"github.com/anoop-dryad/bridgehead/app/config"
	"github.com/anoop-dryad/bridgehead/app/infra/http/handlers"
	"github.com/anoop-dryad/bridgehead/app/infra/http/middleware"
	"github.com/anoop-dryad/bridgehead/app/infra/http/routes"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	http *http.Server
}

func NewServer(cfg config.App, deps handlers.Dependencies, log *zap.Logger) *Server {
	engine := gin.New()
	engine.Use(middleware.Logger(log))
	engine.Use(gin.Recovery())
	routes.Register(engine, deps, cfg)
	return &Server{
		http: &http.Server{
			Addr:           ":8080",
			Handler:        engine,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
	}

}

func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		if err := s.http.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done(): // OS signal received
		return s.http.Shutdown(context.Background())
	}
}
