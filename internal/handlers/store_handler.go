package handlers

import (
	"context"
	"net/http"
	"time"

	"pos-saas-backend/internal/helpers"
	"pos-saas-backend/internal/repositories"

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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stores, err := h.StoreRepository.FindAll(ctx)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get stores", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Stores retrieved successfully", stores)
}
