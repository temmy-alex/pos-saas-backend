package routes

import (
	"database/sql"

	"pos-saas-backend/internal/handlers"
	"pos-saas-backend/internal/repositories"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, db *sql.DB) {
	healthHandler := handlers.NewHealthHandler()
	databaseHandler := handlers.NewDatabaseHandler(db)

	storeRepository := repositories.NewStoreRepository(db)
	branchRepository := repositories.NewBranchRepository(db)

	storeHandler := handlers.NewStoreHandler(storeRepository)
	branchHandler := handlers.NewBranchHandler(branchRepository)

	router.GET("/health", healthHandler.HealthCheck)

	api := router.Group("/api")
	{
		api.GET("/health", healthHandler.HealthCheck)
		api.GET("/db-ping", databaseHandler.Ping)

		api.GET("/stores", storeHandler.FindAll)
		api.GET("/branches", branchHandler.FindAll)
	}
}
