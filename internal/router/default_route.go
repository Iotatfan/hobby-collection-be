package router

import (
	"github.com/gin-gonic/gin"
	"github.com/iotatfan/hobby-collection-be/internal/common"
)

func SetDefaultRoute(g *gin.Engine) {
	g.NoRoute(common.NoRouteHandler)
	g.Static("/docs", "./dist")
}
