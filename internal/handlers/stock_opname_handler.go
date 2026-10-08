package handlers

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pos-saas-backend/internal/helpers"
	"pos-saas-backend/internal/models"
	"pos-saas-backend/internal/repositories"
	"pos-saas-backend/internal/requests"

	"github.com/gin-gonic/gin"
)

type StockOpnameHandler struct {
	StockOpnameRepository *repositories.StockOpnameRepository
	BranchRepository      *repositories.BranchRepository
}

func NewStockOpnameHandler(
	stockOpnameRepository *repositories.StockOpnameRepository,
	branchRepository *repositories.BranchRepository,
) *StockOpnameHandler {
	return &StockOpnameHandler{
		StockOpnameRepository: stockOpnameRepository,
		BranchRepository:      branchRepository,
	}
}

func (h *StockOpnameHandler) FindAll(c *gin.Context) {
	h.findAll(c, false)
}

func (h *StockOpnameHandler) FindAllMobile(c *gin.Context) {
	h.findAll(c, true)
}

func (h *StockOpnameHandler) findAll(c *gin.Context, mobileContract bool) {
	var filter requests.StockOpnameFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Invalid query parameters", err.Error())
		return
	}

	scope, err := helpers.GetAuthScope(c)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}
	filter.StoreID, err = helpers.ApplyStoreScope(scope, filter.StoreID)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}
	filter.BranchID, err = helpers.ApplyBranchScope(scope, filter.BranchID)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	filter, err = normalizeStockOpnameFilter(filter)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Invalid query parameters", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if mobileContract {
		details, total, err := h.StockOpnameRepository.FindMobileDetails(ctx, filter)
		if err != nil {
			helpers.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data stock opname", err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Data stock opname berhasil diambil",
			"data":    details,
			"pagination": gin.H{
				"page":     filter.Page,
				"per_page": filter.Limit,
				"total":    total,
				"has_more": int64(filter.Page*filter.Limit) < total,
			},
		})
		return
	}

	opnames, total, err := h.StockOpnameRepository.FindAll(ctx, filter)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get stock opnames", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Stock opnames retrieved successfully", gin.H{
		"items": opnames,
		"pagination": gin.H{
			"page":        filter.Page,
			"limit":       filter.Limit,
			"total":       total,
			"total_pages": int64(math.Ceil(float64(total) / float64(filter.Limit))),
		},
		"filters": gin.H{
			"store_id":  filter.StoreID,
			"branch_id": filter.BranchID,
			"status":    filter.Status,
			"search":    filter.Search,
		},
	})
}

func (h *StockOpnameHandler) FindByID(c *gin.Context) {
	id, ok := stockOpnameIDParam(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	opname, err := h.StockOpnameRepository.FindByID(ctx, id)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get stock opname", err.Error())
		return
	}
	if opname == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Stock opname not found", "stock opname data not found")
		return
	}
	if !authorizeStockOpname(c, opname) {
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Stock opname retrieved successfully", opname)
}

func (h *StockOpnameHandler) Create(c *gin.Context) {
	var request requests.StockOpnameRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Invalid request payload", err.Error())
		return
	}

	items := request.NormalizedItems()
	if len(items) == 0 {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Invalid request payload", "at least one item is required")
		return
	}
	for _, item := range items {
		if item.ProductID <= 0 || item.OldStock == nil || item.NewStock < 0 || *item.OldStock < 0 {
			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Invalid request payload", "product_id, old_stock, and new_stock are required and stock values cannot be negative")
			return
		}
		if len(item.Note) > 255 {
			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Invalid request payload", "item note cannot exceed 255 characters")
			return
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	scope, branch, ok := h.resolveBranch(c, ctx, request.BranchID)
	if !ok {
		return
	}

	opname, err := h.StockOpnameRepository.Create(ctx, branch.StoreID, branch.ID, scope.UserID, request)
	if err != nil {
		var conflict *repositories.StockChangedError
		if errors.As(err, &conflict) {
			helpers.ErrorResponse(c, http.StatusConflict, "Stock changed", err.Error())
			return
		}
		helpers.ErrorResponse(c, http.StatusBadRequest, "Failed to create stock opname", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusCreated, "Stock opname draft created successfully", opname)
}

func (h *StockOpnameHandler) Complete(c *gin.Context) {
	h.changeStatus(c, true)
}

func (h *StockOpnameHandler) Cancel(c *gin.Context) {
	h.changeStatus(c, false)
}

func (h *StockOpnameHandler) changeStatus(c *gin.Context, complete bool) {
	id, ok := stockOpnameIDParam(c)
	if !ok {
		return
	}

	var cancelRequest requests.CancelStockOpnameRequest
	if !complete {
		if err := c.ShouldBindJSON(&cancelRequest); err != nil {
			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Invalid request payload", err.Error())
			return
		}
		cancelRequest.Reason = strings.TrimSpace(cancelRequest.Reason)
		if cancelRequest.Reason == "" {
			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Invalid request payload", "reason is required")
			return
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	existing, err := h.StockOpnameRepository.FindByID(ctx, id)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get stock opname", err.Error())
		return
	}
	if existing == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Stock opname not found", "stock opname data not found")
		return
	}
	if !authorizeStockOpname(c, existing) {
		return
	}

	scope, err := helpers.GetAuthScope(c)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	var opname *models.StockOpname
	if complete {
		opname, err = h.StockOpnameRepository.Complete(ctx, id, scope.UserID)
	} else {
		opname, err = h.StockOpnameRepository.Cancel(ctx, id, scope.UserID, cancelRequest.Reason)
	}
	if err != nil {
		var conflict *repositories.StockChangedError
		switch {
		case errors.As(err, &conflict):
			helpers.ErrorResponse(c, http.StatusConflict, "Stock changed", err.Error())
		case errors.Is(err, repositories.ErrStockOpnameNotFound):
			helpers.ErrorResponse(c, http.StatusNotFound, "Stock opname not found", err.Error())
		case errors.Is(err, repositories.ErrStockOpnameAlreadyCompleted),
			errors.Is(err, repositories.ErrStockOpnameAlreadyCancelled):
			helpers.ErrorResponse(c, http.StatusConflict, "Stock opname status conflict", err.Error())
		default:
			helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to update stock opname", err.Error())
		}
		return
	}

	message := "Stock opname cancelled successfully"
	if complete {
		message = "Stock opname completed and product stock updated successfully"
	}
	helpers.SuccessResponse(c, http.StatusOK, message, opname)
}

func (h *StockOpnameHandler) resolveBranch(
	c *gin.Context,
	ctx context.Context,
	requestedBranchID int64,
) (*helpers.AuthScope, *models.Branch, bool) {
	scope, err := helpers.GetAuthScope(c)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return nil, nil, false
	}

	branchID := requestedBranchID
	if branchID <= 0 {
		branchID = scope.BranchID
	}
	if branchID <= 0 {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Invalid request payload", "branch_id is required")
		return nil, nil, false
	}

	branch, err := h.BranchRepository.FindByID(ctx, branchID)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get branch", err.Error())
		return nil, nil, false
	}
	if branch == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Branch not found", "branch data not found")
		return nil, nil, false
	}
	if err := helpers.EnsureStoreAccess(scope, branch.StoreID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return nil, nil, false
	}
	if err := helpers.EnsureBranchAccess(scope, branch.ID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return nil, nil, false
	}

	return scope, branch, true
}

func authorizeStockOpname(c *gin.Context, opname *models.StockOpname) bool {
	scope, err := helpers.GetAuthScope(c)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return false
	}
	if err := helpers.EnsureStoreAccess(scope, opname.StoreID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return false
	}
	if err := helpers.EnsureBranchAccess(scope, opname.BranchID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return false
	}
	return true
}

func normalizeStockOpnameFilter(filter requests.StockOpnameFilterRequest) (requests.StockOpnameFilterRequest, error) {
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))
	if filter.Status != "" && filter.Status != "draft" && filter.Status != "completed" && filter.Status != "cancelled" {
		return filter, errors.New("status must be draft, completed, or cancelled")
	}
	filter.Search = strings.TrimSpace(filter.Search)
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = filter.PerPage
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	return filter, nil
}

func stockOpnameIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid stock opname id", "stock opname id must be a positive integer")
		return 0, false
	}
	return id, true
}
