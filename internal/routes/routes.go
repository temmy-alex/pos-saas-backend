package routes

import (
	"database/sql"

	"pos-saas-backend/internal/config"
	"pos-saas-backend/internal/handlers"
	"pos-saas-backend/internal/middlewares"
	"pos-saas-backend/internal/repositories"
	"pos-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, db *sql.DB) {
	jwtConfig := config.LoadJWTConfig()

	healthHandler := handlers.NewHealthHandler()
	databaseHandler := handlers.NewDatabaseHandler(db)

	storeRepository := repositories.NewStoreRepository(db)
	branchRepository := repositories.NewBranchRepository(db)
	userRepository := repositories.NewUserRepository(db)

	storeHandler := handlers.NewStoreHandler(storeRepository)
	branchHandler := handlers.NewBranchHandler(branchRepository)

	authService := services.NewAuthService(
		userRepository,
		jwtConfig.Secret,
		jwtConfig.AccessTokenExpiresMinutes,
	)

	authHandler := handlers.NewAuthHandler(authService)
	authMiddleware := middlewares.NewAuthMiddleware(jwtConfig.Secret)
	roleMiddleware := middlewares.NewRoleMiddleware()

	router.GET("/health", healthHandler.HealthCheck)

	api := router.Group("/api")
	{
		api.GET("/health", healthHandler.HealthCheck)
		api.GET("/db-ping", databaseHandler.Ping)

		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", authMiddleware.RequireAuth(), authHandler.Me)
		}

		protected := api.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			protected.GET(
				"/stores",
				roleMiddleware.RequireRoles("superadmin", "admin"),
				storeHandler.FindAll,
			)

			protected.GET(
				"/branches",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				branchHandler.FindAll,
			)
		}
	}
}
