package router

import (
	"github.com/gin-gonic/gin"
	"github.com/iotatfan/hobby-collection-be/internal/admin/handler"
)

func SetAdminRoutes(g *gin.Engine, aH handler.AdminHandler) {
	g.POST("/admin/token", aH.IssueToken)
}
