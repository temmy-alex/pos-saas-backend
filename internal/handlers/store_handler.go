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

type StoreHandler struct {
	StoreRepository *repositories.StoreRepository
}

func NewStoreHandler(storeRepository *repositories.StoreRepository) *StoreHandler {
	return &StoreHandler{
		StoreRepository: storeRepository,
	}
}

func (h *StoreHandler) FindAll(c *gin.Context) {
	scope, err := helpers.GetAuthScope(c)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	storeID, err := helpers.ApplyStoreScope(scope, 0)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stores, err := h.StoreRepository.FindAll(ctx, storeID)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get stores", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Stores retrieved successfully", stores)
}

func (h *StoreHandler) FindByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid store id", err.Error())
		return
	}

	scope, err := helpers.GetAuthScope(c)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	if err := helpers.EnsureStoreAccess(scope, id); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store, err := h.StoreRepository.FindByID(ctx, id)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get store", err.Error())
		return
	}

	if store == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Store not found", "store data not found")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Store retrieved successfully", store)
}

func (h *StoreHandler) Create(c *gin.Context) {
	var request requests.StoreRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store, err := h.StoreRepository.Create(ctx, request)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to create store", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusCreated, "Store created successfully", store)
}

func (h *StoreHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid store id", err.Error())
		return
	}

	var request requests.StoreRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store, err := h.StoreRepository.Update(ctx, id, request)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to update store", err.Error())
		return
	}

	if store == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Store not found", "store data not found")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Store updated successfully", store)
}

func (h *StoreHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid store id", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	deleted, err := h.StoreRepository.Delete(ctx, id)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete store", err.Error())
		return
	}

	if !deleted {
		helpers.ErrorResponse(c, http.StatusNotFound, "Store not found", "store data not found")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Store deleted successfully", gin.H{
		"id": id,
	})
}
