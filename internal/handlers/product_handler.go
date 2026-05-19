package handlers

import (
	"context"
	"errors"
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	products, err := h.ProductRepository.FindAll(ctx)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get products", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Products retrieved successfully", products)
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

	if err := h.handleProductImage(c, &request); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Failed to upload product image", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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
