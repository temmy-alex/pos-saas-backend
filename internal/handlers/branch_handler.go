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

func (h *BranchHandler) FindByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid branch id", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	branch, err := h.BranchRepository.FindByID(ctx, id)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get branch", err.Error())
		return
	}

	if branch == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Branch not found", "branch data not found")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Branch retrieved successfully", branch)
}

func (h *BranchHandler) Create(c *gin.Context) {
	var request requests.BranchRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	branch, err := h.BranchRepository.Create(ctx, request)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to create branch", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusCreated, "Branch created successfully", branch)
}

func (h *BranchHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid branch id", err.Error())
		return
	}

	var request requests.BranchRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	branch, err := h.BranchRepository.Update(ctx, id, request)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to update branch", err.Error())
		return
	}

	if branch == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Branch not found", "branch data not found")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Branch updated successfully", branch)
}

func (h *BranchHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid branch id", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	deleted, err := h.BranchRepository.Delete(ctx, id)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete branch", err.Error())
		return
	}

	if !deleted {
		helpers.ErrorResponse(c, http.StatusNotFound, "Branch not found", "branch data not found")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Branch deleted successfully", gin.H{
		"id": id,
	})
}
