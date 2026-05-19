package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"pos-saas-backend/internal/helpers"
	"pos-saas-backend/internal/repositories"
	"pos-saas-backend/internal/requests"
	"pos-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	TransactionRepository *repositories.TransactionRepository
}

func NewTransactionHandler(transactionRepository *repositories.TransactionRepository) *TransactionHandler {
	return &TransactionHandler{
		TransactionRepository: transactionRepository,
	}
}

func (h *TransactionHandler) FindAll(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	transactions, err := h.TransactionRepository.FindAll(ctx)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get transactions", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Transactions retrieved successfully", transactions)
}

func (h *TransactionHandler) FindByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid transaction id", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	transaction, err := h.TransactionRepository.FindByID(ctx, id)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get transaction", err.Error())
		return
	}

	if transaction == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Transaction not found", "transaction data not found")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Transaction retrieved successfully", transaction)
}

func (h *TransactionHandler) Create(c *gin.Context) {
	var request requests.TransactionRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	authUserValue, exists := c.Get("auth_user")
	if !exists {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "auth user not found")
		return
	}

	claims, ok := authUserValue.(*services.AuthClaims)
	if !ok {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "invalid auth user claims")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	transaction, err := h.TransactionRepository.Create(ctx, claims.UserID, request)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Failed to create transaction", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusCreated, "Transaction created successfully", transaction)
}
