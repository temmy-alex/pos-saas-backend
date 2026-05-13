package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"pos-saas-backend/internal/helpers"
	"pos-saas-backend/internal/repositories"
	"pos-saas-backend/internal/requests"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	CategoryRepository *repositories.CategoryRepository
}

func NewCategoryHandler(categoryRepository *repositories.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{
		CategoryRepository: categoryRepository,
	}
}

func (h *CategoryHandler) FindAll(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	categories, err := h.CategoryRepository.FindAll(ctx)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get categories", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Categories retrieved successfully", categories)
}

func (h *CategoryHandler) FindByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid category id", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	category, err := h.CategoryRepository.FindByID(ctx, id)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get category", err.Error())
		return
	}

	if category == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Category not found", "category data not found")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Category retrieved successfully", category)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var request requests.CategoryRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	category, err := h.CategoryRepository.Create(ctx, request)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to create category", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusCreated, "Category created successfully", category)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid category id", err.Error())
		return
	}

	var request requests.CategoryRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	category, err := h.CategoryRepository.Update(ctx, id, request)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to update category", err.Error())
		return
	}

	if category == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Category not found", "category data not found")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Category updated successfully", category)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid category id", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	deleted, err := h.CategoryRepository.Delete(ctx, id)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete category", err.Error())
		return
	}

	if !deleted {
		helpers.ErrorResponse(c, http.StatusNotFound, "Category not found", "category data not found")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Category deleted successfully", gin.H{
		"id": id,
	})
}
