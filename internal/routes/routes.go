package routes

import (
	"database/sql"

	"pos-saas-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, db *sql.DB) {
	healthHandler := handlers.NewHealthHandler()
	databaseHandler := handlers.NewDatabaseHandler(db)

	router.GET("/health", healthHandler.HealthCheck)

	api := router.Group("/api")
	{
		api.GET("/health", healthHandler.HealthCheck)
		api.GET("/db-ping", databaseHandler.Ping)
	}
}
