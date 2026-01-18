package app

import (
	"net/http"
	"time"

	"github.com/anoop-dryad/bridgehead/cmd/api/app/routers"
	"github.com/gin-gonic/gin"
)

func InitServer() {
	engine := gin.Default()
	RegisterRoutes(engine)
	server := &http.Server{
		Addr:           ":8080",
		Handler:        engine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	server.ListenAndServe()
}

func RegisterRoutes(engine *gin.Engine) *gin.Engine {

	// base of all api's will start with http(s)://{host}/api/
	api := engine.Group("/api")
	// version v1 : http(s)://{host}/api/v1
	v1 := api.Group("/v1")
	{
		// health group : http(s)://{host}/api/v1/
		health := v1.Group("/")
		routers.Health(health)
	}

	return engine
}
