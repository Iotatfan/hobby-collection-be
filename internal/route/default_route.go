package route

import (
	"github.com/gin-gonic/gin"
	"github.com/iotatfan/hobby-collection-be/internal/helper"
)

func SetDefaultRoute(g *gin.Engine) {
	g.NoRoute(helper.NoRouteHandler)
	g.Static("/docs", "./dist")
}
