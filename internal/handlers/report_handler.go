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

type ReportHandler struct {
	ReportRepository *repositories.ReportRepository
}

func NewReportHandler(reportRepository *repositories.ReportRepository) *ReportHandler {
	return &ReportHandler{
		ReportRepository: reportRepository,
	}
}

func (h *ReportHandler) DailySales(c *gin.Context) {
	var request requests.DailySalesReportRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
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

	summary, paymentBreakdown, topProducts, err := h.ReportRepository.GetDailySalesReport(ctx, request)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get daily sales report", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Daily sales report retrieved successfully", gin.H{
		"summary":           summary,
		"payment_breakdown": paymentBreakdown,
		"top_products":      topProducts,
		"filters": gin.H{
			"date":      request.Date,
			"store_id":  request.StoreID,
			"branch_id": request.BranchID,
		},
	})
}
