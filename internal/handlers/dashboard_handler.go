package handlers

import (
	"context"
	"net/http"
	"time"

	"pos-saas-backend/internal/helpers"
	"pos-saas-backend/internal/repositories"
	"pos-saas-backend/internal/requests"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	DashboardRepository *repositories.DashboardRepository
}

func NewDashboardHandler(dashboardRepository *repositories.DashboardRepository) *DashboardHandler {
	return &DashboardHandler{
		DashboardRepository: dashboardRepository,
	}
}

func (h *DashboardHandler) Index(c *gin.Context) {
	var request requests.DashboardRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	if request.Date == "" {
		location, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			location = time.FixedZone("Asia/Jakarta", 7*60*60)
		}

		request.Date = time.Now().In(location).Format("2006-01-02")
	}

	if _, err := time.Parse("2006-01-02", request.Date); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid date format", "date format must be YYYY-MM-DD")
		return
	}

	scope, err := helpers.GetAuthScope(c)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	storeID, err := helpers.ApplyStoreScope(scope, request.StoreID)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	branchID, err := helpers.ApplyBranchScope(scope, request.BranchID)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	request.StoreID = storeID
	request.BranchID = branchID

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dashboard, err := h.DashboardRepository.GetDashboard(ctx, request)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get dashboard", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Dashboard retrieved successfully", gin.H{
		"summary":             dashboard.Summary,
		"sales_per_hour":      dashboard.SalesPerHour,
		"low_stock_products":  dashboard.LowStockProducts,
		"recent_transactions": dashboard.RecentTransactions,
		"filters": gin.H{
			"date":      request.Date,
			"store_id":  request.StoreID,
			"branch_id": request.BranchID,
		},
	})
}
