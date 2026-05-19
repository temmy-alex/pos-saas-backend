package handlers

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"pos-saas-backend/internal/helpers"
	"pos-saas-backend/internal/repositories"
	"pos-saas-backend/internal/requests"
	"pos-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	ProductRepository   *repositories.ProductRepository
	LocalStorageService *services.LocalStorageService
}

func NewProductHandler(
	productRepository *repositories.ProductRepository,
	localStorageService *services.LocalStorageService,
) *ProductHandler {
	return &ProductHandler{
		ProductRepository:   productRepository,
		LocalStorageService: localStorageService,
	}
}

func (h *ProductHandler) FindAll(c *gin.Context) {
	var filter requests.ProductFilterRequest

	if err := c.ShouldBindQuery(&filter); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return
	}

	scope, err := helpers.GetAuthScope(c)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	storeID, err := helpers.ApplyStoreScope(scope, filter.StoreID)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	branchID, err := helpers.ApplyBranchScope(scope, filter.BranchID)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	filter.StoreID = storeID
	filter.BranchID = branchID
	filter = normalizeProductFilter(filter)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	products, total, err := h.ProductRepository.FindAll(ctx, filter)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get products", err.Error())
		return
	}

	totalPages := int64(0)
	if filter.Limit > 0 {
		totalPages = int64(math.Ceil(float64(total) / float64(filter.Limit)))
	}

	helpers.SuccessResponse(c, http.StatusOK, "Products retrieved successfully", gin.H{
		"items": products,
		"pagination": gin.H{
			"page":        filter.Page,
			"limit":       filter.Limit,
			"total":       total,
			"total_pages": totalPages,
		},
		"filters": gin.H{
			"store_id":    filter.StoreID,
			"branch_id":   filter.BranchID,
			"category_id": filter.CategoryID,
			"search":      filter.Search,
			"is_active":   filter.IsActive,
		},
	})
}

func (h *ProductHandler) FindByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid product id", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	product, err := h.ProductRepository.FindByID(ctx, id)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get product", err.Error())
		return
	}

	if product == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Product not found", "product data not found")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Product retrieved successfully", product)
}

func (h *ProductHandler) Create(c *gin.Context) {
	var request requests.ProductRequest

	if err := c.ShouldBind(&request); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	scope, err := helpers.GetAuthScope(c)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	if err := helpers.EnsureStoreAccess(scope, request.StoreID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	if err := helpers.EnsureBranchAccess(scope, request.BranchID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	if err := h.handleProductImage(c, &request); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Failed to upload product image", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	product, err := h.ProductRepository.Create(ctx, request)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to create product", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusCreated, "Product created successfully", product)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid product id", err.Error())
		return
	}

	var request requests.ProductRequest

	if err := c.ShouldBind(&request); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	scope, err := helpers.GetAuthScope(c)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	if err := helpers.EnsureStoreAccess(scope, request.StoreID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	if err := helpers.EnsureBranchAccess(scope, request.BranchID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	if err := h.handleProductImage(c, &request); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Failed to upload product image", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	existingProduct, err := h.ProductRepository.FindByID(ctx, id)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get product", err.Error())
		return
	}

	if existingProduct == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Product not found", "product data not found")
		return
	}

	if err := helpers.EnsureStoreAccess(scope, existingProduct.StoreID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	if err := helpers.EnsureBranchAccess(scope, existingProduct.BranchID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	product, err := h.ProductRepository.Update(ctx, id, request)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to update product", err.Error())
		return
	}

	if product == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Product not found", "product data not found")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Product updated successfully", product)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid product id", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	deleted, err := h.ProductRepository.Delete(ctx, id)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete product", err.Error())
		return
	}

	if !deleted {
		helpers.ErrorResponse(c, http.StatusNotFound, "Product not found", "product data not found")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Product deleted successfully", gin.H{
		"id": id,
	})
}

func (h *ProductHandler) handleProductImage(c *gin.Context, request *requests.ProductRequest) error {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return nil
		}

		return nil
	}

	uploadedFile, err := h.LocalStorageService.UploadProductImage(fileHeader)
	if err != nil {
		return err
	}

	if uploadedFile == nil {
		return nil
	}

	request.ImagePath = &uploadedFile.Path
	request.ImageURL = &uploadedFile.URL
	request.ImageDisk = &uploadedFile.Disk

	return nil
}

func normalizeProductFilter(filter requests.ProductFilterRequest) requests.ProductFilterRequest {
	if filter.Page <= 0 {
		filter.Page = 1
	}

	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	if filter.Limit > 100 {
		filter.Limit = 100
	}

	return filter
}
