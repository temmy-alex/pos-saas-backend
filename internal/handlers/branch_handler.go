package handlers

import (
	"context"
	"net/http"
	"time"

	"pos-saas-backend/internal/helpers"
	"pos-saas-backend/internal/repositories"

	"github.com/gin-gonic/gin"
)

type BranchHandler struct {
	BranchRepository *repositories.BranchRepository
}

func NewBranchHandler(branchRepository *repositories.BranchRepository) *BranchHandler {
	return &BranchHandler{
		BranchRepository: branchRepository,
	}
}

func (h *BranchHandler) FindAll(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	branches, err := h.BranchRepository.FindAll(ctx)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get branches", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Branches retrieved successfully", branches)
}
