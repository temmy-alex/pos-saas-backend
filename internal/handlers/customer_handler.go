package handlers

import (
	"context"
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

type CustomerHandler struct {
	CustomerRepository *repositories.CustomerRepository
	BranchRepository   *repositories.BranchRepository
}

func NewCustomerHandler(
	customerRepository *repositories.CustomerRepository,
	branchRepository *repositories.BranchRepository,
) *CustomerHandler {
	return &CustomerHandler{
		CustomerRepository: customerRepository,
		BranchRepository:   branchRepository,
	}
}

func (h *CustomerHandler) FindAll(c *gin.Context) {
	h.findAll(c, false)
}

func (h *CustomerHandler) FindAllMobile(c *gin.Context) {
	h.findAll(c, true)
}

func (h *CustomerHandler) findAll(c *gin.Context, mobileContract bool) {
	var filter requests.CustomerFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err.Error())
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

	filter = normalizeCustomerFilter(filter)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	customers, total, err := h.CustomerRepository.FindAll(ctx, filter)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get customers", err.Error())
		return
	}

	if mobileContract {
		c.JSON(http.StatusOK, gin.H{
			"message": "Data customer berhasil diambil",
			"data":    customers,
			"pagination": gin.H{
				"page":     filter.Page,
				"per_page": filter.Limit,
				"total":    total,
				"has_more": int64(filter.Page*filter.Limit) < total,
			},
		})
		return
	}

	totalPages := int64(math.Ceil(float64(total) / float64(filter.Limit)))
	helpers.SuccessResponse(c, http.StatusOK, "Customers retrieved successfully", gin.H{
		"items": customers,
		"pagination": gin.H{
			"page":        filter.Page,
			"limit":       filter.Limit,
			"total":       total,
			"total_pages": totalPages,
		},
		"filters": gin.H{
			"store_id":  filter.StoreID,
			"branch_id": filter.BranchID,
			"search":    filter.Search,
		},
	})
}

func (h *CustomerHandler) FindByID(c *gin.Context) {
	id, ok := customerIDParam(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	customer, err := h.CustomerRepository.FindByID(ctx, id)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get customer", err.Error())
		return
	}
	if customer == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Customer not found", "customer data not found")
		return
	}

	if !h.authorizeCustomer(c, customer) {
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Customer retrieved successfully", customer)
}

func (h *CustomerHandler) Create(c *gin.Context) {
	h.create(c, false)
}

func (h *CustomerHandler) CreateMobile(c *gin.Context) {
	h.create(c, true)
}

func (h *CustomerHandler) create(c *gin.Context, mobileContract bool) {
	var request requests.CustomerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Invalid request payload", err.Error())
		return
	}
	if strings.TrimSpace(request.Name) == "" || strings.TrimSpace(request.Phone) == "" {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Invalid request payload", "name and phone are required")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	storeID, branchID, ok := h.resolveCustomerBranch(c, ctx, request.BranchID)
	if !ok {
		return
	}

	customer, err := h.CustomerRepository.Create(ctx, storeID, branchID, request)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to create customer", err.Error())
		return
	}

	if mobileContract {
		c.JSON(http.StatusCreated, gin.H{
			"message": "Customer berhasil dibuat",
			"data":    customer,
		})
		return
	}

	helpers.SuccessResponse(c, http.StatusCreated, "Customer created successfully", customer)
}

func (h *CustomerHandler) Update(c *gin.Context) {
	id, ok := customerIDParam(c)
	if !ok {
		return
	}

	var request requests.CustomerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Invalid request payload", err.Error())
		return
	}
	if strings.TrimSpace(request.Name) == "" || strings.TrimSpace(request.Phone) == "" {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Invalid request payload", "name and phone are required")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	existing, err := h.CustomerRepository.FindByID(ctx, id)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get customer", err.Error())
		return
	}
	if existing == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Customer not found", "customer data not found")
		return
	}
	if !h.authorizeCustomer(c, existing) {
		return
	}

	branchID := request.BranchID
	if branchID <= 0 {
		branchID = existing.BranchID
	}
	storeID, branchID, ok := h.resolveCustomerBranch(c, ctx, branchID)
	if !ok {
		return
	}

	customer, err := h.CustomerRepository.Update(ctx, id, storeID, branchID, request)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to update customer", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Customer updated successfully", customer)
}

func (h *CustomerHandler) Delete(c *gin.Context) {
	id, ok := customerIDParam(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	customer, err := h.CustomerRepository.FindByID(ctx, id)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get customer", err.Error())
		return
	}
	if customer == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Customer not found", "customer data not found")
		return
	}
	if !h.authorizeCustomer(c, customer) {
		return
	}

	deleted, err := h.CustomerRepository.Delete(ctx, id)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete customer", err.Error())
		return
	}
	if !deleted {
		helpers.ErrorResponse(c, http.StatusNotFound, "Customer not found", "customer data not found")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Customer deleted successfully", gin.H{"id": id})
}

func (h *CustomerHandler) resolveCustomerBranch(c *gin.Context, ctx context.Context, requestedBranchID int64) (int64, int64, bool) {
	scope, err := helpers.GetAuthScope(c)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return 0, 0, false
	}

	branchID := requestedBranchID
	if branchID <= 0 {
		branchID = scope.BranchID
	}
	if branchID <= 0 {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Invalid request payload", "branch_id is required")
		return 0, 0, false
	}

	branch, err := h.BranchRepository.FindByID(ctx, branchID)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get branch", err.Error())
		return 0, 0, false
	}
	if branch == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Branch not found", "branch data not found")
		return 0, 0, false
	}

	if err := helpers.EnsureStoreAccess(scope, branch.StoreID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return 0, 0, false
	}
	if err := helpers.EnsureBranchAccess(scope, branch.ID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return 0, 0, false
	}

	return branch.StoreID, branch.ID, true
}

func (h *CustomerHandler) authorizeCustomer(c *gin.Context, customer *models.Customer) bool {
	scope, err := helpers.GetAuthScope(c)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return false
	}
	if err := helpers.EnsureStoreAccess(scope, customer.StoreID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return false
	}
	if err := helpers.EnsureBranchAccess(scope, customer.BranchID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return false
	}
	return true
}

func customerIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid customer id", "customer id must be a positive integer")
		return 0, false
	}
	return id, true
}

func normalizeCustomerFilter(filter requests.CustomerFilterRequest) requests.CustomerFilterRequest {
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
	filter.Search = strings.TrimSpace(filter.Search)
	return filter
}
