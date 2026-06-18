package swagger

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Register(router *gin.RouterGroup) {
	router.GET("/swagger/v1/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
