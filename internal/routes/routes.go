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
	categoryRepository := repositories.NewCategoryRepository(db)
	userRepository := repositories.NewUserRepository(db)

	storeHandler := handlers.NewStoreHandler(storeRepository)
	branchHandler := handlers.NewBranchHandler(branchRepository)
	categoryHandler := handlers.NewCategoryHandler(categoryRepository)

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

			protected.POST(
				"/stores",
				roleMiddleware.RequireRoles("superadmin"),
				storeHandler.Create,
			)

			protected.GET(
				"/stores/:id",
				roleMiddleware.RequireRoles("superadmin", "admin"),
				storeHandler.FindByID,
			)

			protected.PUT(
				"/stores/:id",
				roleMiddleware.RequireRoles("superadmin"),
				storeHandler.Update,
			)

			protected.DELETE(
				"/stores/:id",
				roleMiddleware.RequireRoles("superadmin"),
				storeHandler.Delete,
			)

			protected.GET(
				"/branches",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				branchHandler.FindAll,
			)

			protected.POST(
				"/branches",
				roleMiddleware.RequireRoles("superadmin"),
				branchHandler.Create,
			)

			protected.GET(
				"/branches/:id",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				branchHandler.FindByID,
			)

			protected.PUT(
				"/branches/:id",
				roleMiddleware.RequireRoles("superadmin"),
				branchHandler.Update,
			)

			protected.DELETE(
				"/branches/:id",
				roleMiddleware.RequireRoles("superadmin"),
				branchHandler.Delete,
			)

			protected.GET(
				"/categories",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				categoryHandler.FindAll,
			)

			protected.POST(
				"/categories",
				roleMiddleware.RequireRoles("superadmin", "admin"),
				categoryHandler.Create,
			)

			protected.GET(
				"/categories/:id",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				categoryHandler.FindByID,
			)

			protected.PUT(
				"/categories/:id",
				roleMiddleware.RequireRoles("superadmin", "admin"),
				categoryHandler.Update,
			)

			protected.DELETE(
				"/categories/:id",
				roleMiddleware.RequireRoles("superadmin", "admin"),
				categoryHandler.Delete,
			)
		}
	}
}
