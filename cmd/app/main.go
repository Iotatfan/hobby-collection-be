package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iotatfan/hobby-collection-be/internal/collection/entity"
	"github.com/iotatfan/hobby-collection-be/internal/config"
	"github.com/iotatfan/hobby-collection-be/internal/handle"
	"github.com/iotatfan/hobby-collection-be/internal/middleware"
	"github.com/iotatfan/hobby-collection-be/internal/route"
	"github.com/iotatfan/hobby-collection-be/pkg/database/gorm"
)

func handleRequests(cfg *config.Config) {
	db := gorm.NewDB(&cfg.Postgres)
	// cld := cloud.NewCld(&cfg.Cloudinary)
	db.AutoMigrate(&entity.Collection{}, &entity.Grade{}, &entity.ReleaseType{}, &entity.Series{}, &entity.Picture{})

	g := gin.Default()
	g.Use(middleware.CORS())

	route.SetDefaultRoute(g)
	handle.SetupCollection(g, db)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: g,
	}

	go func() {
		// service connections
		log.Printf("listening at port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}

	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	// catching ctx.Done(). timeout of 1 seconds.
	select {
	case <-ctx.Done():
		log.Println("timeout of 1 seconds.")
	}
	log.Println("Server exiting")

}

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		fmt.Println("Config Error: ", err.Error())
	}

	handleRequests(cfg)
}
