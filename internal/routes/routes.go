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
	appConfig := config.LoadAppConfig()
	jwtConfig := config.LoadJWTConfig()

	healthHandler := handlers.NewHealthHandler()
	databaseHandler := handlers.NewDatabaseHandler(db)

	storeRepository := repositories.NewStoreRepository(db)
	branchRepository := repositories.NewBranchRepository(db)
	categoryRepository := repositories.NewCategoryRepository(db)
	productRepository := repositories.NewProductRepository(db)
	customerRepository := repositories.NewCustomerRepository(db)
	storeStatusRepository := repositories.NewStoreStatusRepository(db)
	stockOpnameRepository := repositories.NewStockOpnameRepository(db)
	transactionRepository := repositories.NewTransactionRepository(db)
	reportRepository := repositories.NewReportRepository(db)
	dashboardRepository := repositories.NewDashboardRepository(db)
	userRepository := repositories.NewUserRepository(db)

	storeHandler := handlers.NewStoreHandler(storeRepository)
	branchHandler := handlers.NewBranchHandler(branchRepository)
	categoryHandler := handlers.NewCategoryHandler(categoryRepository)

	localStorageService := services.NewLocalStorageService(appConfig.AppBaseURL)
	productHandler := handlers.NewProductHandler(productRepository, localStorageService)
	customerHandler := handlers.NewCustomerHandler(customerRepository, branchRepository)
	storeStatusHandler := handlers.NewStoreStatusHandler(storeStatusRepository, branchRepository)
	stockOpnameHandler := handlers.NewStockOpnameHandler(stockOpnameRepository, branchRepository)
	transactionHandler := handlers.NewTransactionHandler(transactionRepository)
	reportHandler := handlers.NewReportHandler(reportRepository)
	dashboardHandler := handlers.NewDashboardHandler(dashboardRepository)

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
				"/master/stock-opnames",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				stockOpnameHandler.FindAllMobile,
			)

			protected.GET(
				"/stock-opnames",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				stockOpnameHandler.FindAll,
			)

			protected.POST(
				"/stock-opnames",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				stockOpnameHandler.Create,
			)

			protected.GET(
				"/stock-opnames/:id",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				stockOpnameHandler.FindByID,
			)

			protected.POST(
				"/stock-opnames/:id/complete",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				stockOpnameHandler.Complete,
			)

			protected.POST(
				"/stock-opnames/:id/cancel",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				stockOpnameHandler.Cancel,
			)

			protected.GET(
				"/store/status",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				storeStatusHandler.Status,
			)

			protected.POST(
				"/store/open",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				storeStatusHandler.Open,
			)

			protected.POST(
				"/store/close",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				storeStatusHandler.Close,
			)

			protected.GET(
				"/master/customers",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				customerHandler.FindAllMobile,
			)

			protected.POST(
				"/master/customers",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				customerHandler.CreateMobile,
			)

			protected.GET(
				"/dashboard",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				dashboardHandler.Index,
			)

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

			protected.GET(
				"/products",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				productHandler.FindAll,
			)

			protected.POST(
				"/products",
				roleMiddleware.RequireRoles("superadmin", "admin"),
				productHandler.Create,
			)

			protected.GET(
				"/products/:id",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				productHandler.FindByID,
			)

			protected.PUT(
				"/products/:id",
				roleMiddleware.RequireRoles("superadmin", "admin"),
				productHandler.Update,
			)

			protected.DELETE(
				"/products/:id",
				roleMiddleware.RequireRoles("superadmin", "admin"),
				productHandler.Delete,
			)

			protected.GET(
				"/customers",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				customerHandler.FindAll,
			)

			protected.POST(
				"/customers",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				customerHandler.Create,
			)

			protected.GET(
				"/customers/:id",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				customerHandler.FindByID,
			)

			protected.PUT(
				"/customers/:id",
				roleMiddleware.RequireRoles("superadmin", "admin"),
				customerHandler.Update,
			)

			protected.DELETE(
				"/customers/:id",
				roleMiddleware.RequireRoles("superadmin", "admin"),
				customerHandler.Delete,
			)

			protected.GET(
				"/transactions",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				transactionHandler.FindAll,
			)

			protected.POST(
				"/transactions",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				transactionHandler.Create,
			)

			protected.POST(
				"/transactions/:id/void",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				transactionHandler.Void,
			)

			protected.GET(
				"/transactions/:id/receipt",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				transactionHandler.Receipt,
			)

			protected.GET(
				"/transactions/:id",
				roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
				transactionHandler.FindByID,
			)

			reports := protected.Group("/reports")
			{
				reports.GET(
					"/daily-sales",
					roleMiddleware.RequireRoles("superadmin", "admin", "cashier"),
					reportHandler.DailySales,
				)
			}
		}
	}
}
