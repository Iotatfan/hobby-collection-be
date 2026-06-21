package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/iotatfan/hobby-collection-be/internal/admin/handler"
	"github.com/iotatfan/hobby-collection-be/internal/router"
)

func Register(g *gin.Engine) {
	adminHandler := handler.NewAdminHandler()
	router.SetAdminRoutes(g, adminHandler)
}
