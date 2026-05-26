package handlers

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"time"

	"pos-saas-backend/internal/helpers"
	"pos-saas-backend/internal/models"
	"pos-saas-backend/internal/repositories"
	"pos-saas-backend/internal/requests"
	"pos-saas-backend/internal/responses"

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
	var filter requests.TransactionFilterRequest

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
	filter = normalizeTransactionFilter(filter)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	transactions, total, err := h.TransactionRepository.FindAll(ctx, filter)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get transactions", err.Error())
		return
	}

	totalPages := int64(0)
	if filter.Limit > 0 {
		totalPages = int64(math.Ceil(float64(total) / float64(filter.Limit)))
	}

	helpers.SuccessResponse(c, http.StatusOK, "Transactions retrieved successfully", gin.H{
		"items": transactions,
		"pagination": gin.H{
			"page":        filter.Page,
			"limit":       filter.Limit,
			"total":       total,
			"total_pages": totalPages,
		},
		"filters": gin.H{
			"store_id":       filter.StoreID,
			"branch_id":      filter.BranchID,
			"cashier_id":     filter.CashierID,
			"date":           filter.Date,
			"payment_method": filter.PaymentMethod,
			"search":         filter.Search,
		},
	})
}

func (h *TransactionHandler) FindByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid transaction id", err.Error())
		return
	}

	scope, err := helpers.GetAuthScope(c)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", err.Error())
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

	if err := helpers.EnsureStoreAccess(scope, transaction.StoreID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	if err := helpers.EnsureBranchAccess(scope, transaction.BranchID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Transaction retrieved successfully", transaction)
}

func (h *TransactionHandler) Receipt(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid transaction id", err.Error())
		return
	}

	scope, err := helpers.GetAuthScope(c)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	transaction, err := h.TransactionRepository.FindByID(ctx, id)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get transaction receipt", err.Error())
		return
	}

	if transaction == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Transaction not found", "transaction data not found")
		return
	}

	if err := helpers.EnsureStoreAccess(scope, transaction.StoreID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	if err := helpers.EnsureBranchAccess(scope, transaction.BranchID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	receipt := buildReceiptResponse(transaction)

	helpers.SuccessResponse(c, http.StatusOK, "Transaction receipt retrieved successfully", receipt)
}

func (h *TransactionHandler) Create(c *gin.Context) {
	var request requests.TransactionRequest

	if err := c.ShouldBindJSON(&request); err != nil {
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	transaction, err := h.TransactionRepository.Create(ctx, scope.UserID, request)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Failed to create transaction", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusCreated, "Transaction created successfully", transaction)
}

func (h *TransactionHandler) Void(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid transaction id", err.Error())
		return
	}

	var request requests.VoidTransactionRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	scope, err := helpers.GetAuthScope(c)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	existingTransaction, err := h.TransactionRepository.FindByID(ctx, id)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to get transaction", err.Error())
		return
	}

	if existingTransaction == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Transaction not found", "transaction data not found")
		return
	}

	if err := helpers.EnsureStoreAccess(scope, existingTransaction.StoreID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	if err := helpers.EnsureBranchAccess(scope, existingTransaction.BranchID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	transaction, err := h.TransactionRepository.Void(ctx, id, scope.UserID, request.Reason)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Failed to void transaction", err.Error())
		return
	}

	if transaction == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Transaction not found", "transaction data not found")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Transaction voided successfully", transaction)
}

func normalizeTransactionFilter(filter requests.TransactionFilterRequest) requests.TransactionFilterRequest {
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

func buildReceiptResponse(transaction *models.Transaction) responses.ReceiptResponse {
	items := make([]responses.ReceiptItem, 0)

	for _, item := range transaction.Items {
		grossTotal := item.Price * float64(item.Qty)

		items = append(items, responses.ReceiptItem{
			ID:          item.ID,
			ProductID:   item.ProductID,
			ProductName: item.ProductNameSnapshot,
			ProductSKU:  item.ProductSKUSnapshot,
			Qty:         item.Qty,
			Price:       item.Price,
			GrossTotal:  grossTotal,
			Discount:    item.Discount,
			Subtotal:    item.Subtotal,
		})
	}

	var voidInfo *responses.ReceiptVoid

	if transaction.Status == "void" {
		voidInfo = &responses.ReceiptVoid{
			VoidReason:   transaction.VoidReason,
			VoidedAt:     transaction.VoidedAt,
			VoidedBy:     transaction.VoidedBy,
			VoidedByName: transaction.VoidedByName,
		}
	}

	return responses.ReceiptResponse{
		Store: responses.ReceiptStore{
			ID:   transaction.StoreID,
			Name: transaction.StoreName,
		},
		Branch: responses.ReceiptBranch{
			ID:   transaction.BranchID,
			Name: transaction.BranchName,
		},
		Cashier: responses.ReceiptCashier{
			ID:   transaction.CashierID,
			Name: transaction.CashierName,
		},
		Transaction: responses.ReceiptTransaction{
			ID:                transaction.ID,
			TransactionNumber: transaction.TransactionNumber,
			TransactionDate:   transaction.TransactionDate,
			CustomerName:      transaction.CustomerName,
			Status:            transaction.Status,
			Notes:             transaction.Notes,
		},
		Items: items,
		Summary: responses.ReceiptSummary{
			Subtotal:      transaction.Subtotal,
			DiscountTotal: transaction.DiscountTotal,
			GrandTotal:    transaction.GrandTotal,
		},
		Payment: responses.ReceiptPayment{
			PaymentMethod:  transaction.PaymentMethod,
			CashAmount:     transaction.CashAmount,
			TransferAmount: transaction.TransferAmount,
			PaidAmount:     transaction.CashAmount + transaction.TransferAmount,
			ChangeAmount:   transaction.ChangeAmount,
		},
		Void: voidInfo,
	}
}
