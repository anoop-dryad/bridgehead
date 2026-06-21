package swagger

import (
	"github.com/anoop-dryad/bridgehead/infra/http/swagger/docs"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Register(router *gin.Engine, version string) {
	docs.SwaggerInfo.BasePath = "/" + version // "/v1" or "/v2"

	router.GET("/swagger/"+version+"/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/swagger/"+version+"/doc.json"),
	))
}
