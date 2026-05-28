package main

import (
	"fmt"
	"log"
	"time"

	"pos-saas-backend/internal/config"
	"pos-saas-backend/internal/database"
	"pos-saas-backend/internal/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	appConfig := config.LoadAppConfig()
	databaseConfig := config.LoadDatabaseConfig()

	if appConfig.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := database.NewPostgresConnection(databaseConfig)
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}
	defer db.Close()

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
		ExposeHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Static("/uploads", "./uploads")

	routes.RegisterRoutes(router, db)

	serverAddress := fmt.Sprintf(":%s", appConfig.AppPort)

	log.Println("======================================")
	log.Println("Application :", appConfig.AppName)
	log.Println("Environment :", appConfig.AppEnv)
	log.Println("Port        :", appConfig.AppPort)
	log.Println("Base URL    :", appConfig.AppBaseURL)
	log.Println("Database    :", databaseConfig.Database)
	log.Println("DB Host     :", databaseConfig.Host)
	log.Println("======================================")

	if err := router.Run(serverAddress); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
