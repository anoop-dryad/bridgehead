package app

import (
	"net/http"
	"time"

	"github.com/anoop-dryad/bridgehead/cmd/api/app/routers"
	"github.com/anoop-dryad/bridgehead/cmd/api/docs"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitServer() {
	engine := gin.Default()
	RegisterRoutes(engine)
	RegisterSwagger(engine)
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

	api := engine.Group("/api")
	v1 := api.Group("/v1")
	{
		health := v1.Group("/")
		routers.Health(health)
	}

	return engine
}

func RegisterSwagger(engine *gin.Engine) {
	docs.SwaggerInfo.BasePath = "/api/v1"
	engine.GET("/swagger/v1/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
