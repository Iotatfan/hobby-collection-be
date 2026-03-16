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
	dbgorm "github.com/iotatfan/hobby-collection-be/pkg/database/gorm"
	"github.com/iotatfan/hobby-collection-be/pkg/storage/cloud"
	"gorm.io/gorm"
)

func handleRequests() {
	db := dbgorm.NewDB(&config.GetConfig().Postgres)
	cld := cloud.NewCld(&config.GetConfig().Cloudinary)
	if err := runSchemaMigrations(db); err != nil {
		log.Fatalf("schema migration failed: %v", err)
	}
	if err := db.AutoMigrate(
		&entity.Scale{},
		&entity.Grade{},
		&entity.CollectionType{},
		&entity.ReleaseType{},
		&entity.Manufacturer{},
		&entity.Series{},
		&entity.Collection{},
		&entity.Picture{},
	); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	g := gin.Default()
	g.Use(middleware.CORS())

	route.SetDefaultRoute(g)
	handle.SetupCollection(g, db, cld)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.GetConfig().Server.Port),
		Handler: g,
	}

	go func() {
		// service connections
		log.Printf("listening at port %d", config.AppConfig.Server.Port)
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

func runSchemaMigrations(db *gorm.DB) error {
	if db.Migrator().HasTable(&entity.Collection{}) {
		// hasTypeID := db.Migrator().HasColumn(&entity.Collection{}, "type_id")
		hasGradeID := db.Migrator().HasColumn(&entity.Collection{}, "grade_id")

		if hasGradeID {
			// if err := db.Exec("UPDATE collections SET grade_id = type_id WHERE grade_id = 0").Error; err != nil {
			// 	return err
			// }

			// Drop FK constraint before dropping column
			if db.Migrator().HasConstraint(&entity.Collection{}, "fk_collections_collection_type") {
				fmt.Println("Dropping Constraint")
				db.Migrator().DropConstraint(&entity.Collection{}, "fk_collections_collection_type")
			}

			if db.Migrator().HasConstraint(&entity.Collection{}, "TypeID") {
				fmt.Println("Dropping Constraint")
				db.Migrator().DropConstraint(&entity.Collection{}, "TypeID")
			}
		}
	}

	return nil
}

func main() {
	err := config.InitConfig()
	if err != nil {
		fmt.Println("Config Error: ", err.Error())
	}

	handleRequests()
}
