package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"pos-saas-backend/internal/helpers"
	"pos-saas-backend/internal/models"
	"pos-saas-backend/internal/repositories"
	"pos-saas-backend/internal/requests"

	"github.com/gin-gonic/gin"
)

type StoreStatusHandler struct {
	StoreStatusRepository *repositories.StoreStatusRepository
	BranchRepository      *repositories.BranchRepository
}

func NewStoreStatusHandler(
	storeStatusRepository *repositories.StoreStatusRepository,
	branchRepository *repositories.BranchRepository,
) *StoreStatusHandler {
	return &StoreStatusHandler{
		StoreStatusRepository: storeStatusRepository,
		BranchRepository:      branchRepository,
	}
}

func (h *StoreStatusHandler) Status(c *gin.Context) {
	var request requests.StoreStatusFilterRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Parameter tidak valid", err.Error())
		return
	}

	businessDate, err := parseBusinessDate(request.Date)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Parameter tidak valid", "date must use YYYY-MM-DD format")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, branch, ok := h.resolveBranch(c, ctx, request.BranchID)
	if !ok {
		return
	}

	status, summary, err := h.StoreStatusRepository.GetOrCreate(
		ctx,
		branch.StoreID,
		branch.ID,
		businessDate,
	)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil status store", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Status store berhasil diambil",
		"data":    storeStatusResponse(status, summary, branch.ID, businessDate),
	})
}

func (h *StoreStatusHandler) Open(c *gin.Context) {
	var request requests.OpenStoreRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Parameter tidak valid", err.Error())
		return
	}
	if request.CashOpen == nil || *request.CashOpen < 0 {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Parameter tidak valid", "cash_open must be greater than or equal to zero")
		return
	}
	if request.ClosedAt != nil {
		value := strings.TrimSpace(*request.ClosedAt)
		if _, err := time.Parse("15:04", value); err != nil {
			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Parameter tidak valid", "closed_at must use HH:MM format")
			return
		}
		request.ClosedAt = &value
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	scope, branch, ok := h.resolveBranch(c, ctx, request.BranchID)
	if !ok {
		return
	}

	businessDate, _ := parseBusinessDate("")
	status, err := h.StoreStatusRepository.Open(
		ctx,
		branch.StoreID,
		branch.ID,
		scope.UserID,
		businessDate,
		*request.CashOpen,
		request.ClosedAt,
	)
	if err != nil {
		if errors.Is(err, repositories.ErrStoreAlreadyOpen) || errors.Is(err, repositories.ErrStoreAlreadyClosed) {
			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Toko tidak dapat dibuka", err.Error())
			return
		}
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuka toko", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Toko berhasil dibuka!",
		"data": gin.H{
			"status":    status,
			"branch_id": branch.ID,
			"date":      businessDate,
		},
	})
}

func (h *StoreStatusHandler) Close(c *gin.Context) {
	var request requests.CloseStoreRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Parameter tidak valid", err.Error())
		return
	}
	if request.CashClose == nil || *request.CashClose < 0 {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Parameter tidak valid", "cash_close must be greater than or equal to zero")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	scope, branch, ok := h.resolveBranch(c, ctx, request.BranchID)
	if !ok {
		return
	}

	businessDate, _ := parseBusinessDate("")
	status, summary, err := h.StoreStatusRepository.Close(
		ctx,
		branch.ID,
		scope.UserID,
		businessDate,
		*request.CashClose,
	)
	if err != nil {
		if errors.Is(err, repositories.ErrStoreNotOpen) || errors.Is(err, repositories.ErrStoreAlreadyClosed) {
			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Toko tidak dapat ditutup", err.Error())
			return
		}
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Gagal menutup toko", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Toko berhasil ditutup!",
		"data":    storeStatusResponse(status, summary, branch.ID, businessDate),
	})
}

func (h *StoreStatusHandler) resolveBranch(
	c *gin.Context,
	ctx context.Context,
	requestedBranchID int64,
) (*helpers.AuthScope, *models.Branch, bool) {
	scope, err := helpers.GetAuthScope(c)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return nil, nil, false
	}

	branchID := requestedBranchID
	if branchID <= 0 {
		branchID = scope.BranchID
	}
	if branchID <= 0 {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Parameter tidak valid", "branch_id is required")
		return nil, nil, false
	}

	branch, err := h.BranchRepository.FindByID(ctx, branchID)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil branch", err.Error())
		return nil, nil, false
	}
	if branch == nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Branch tidak ditemukan", "branch data not found")
		return nil, nil, false
	}
	if err := helpers.EnsureStoreAccess(scope, branch.StoreID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Branch tidak dapat diakses", err.Error())
		return nil, nil, false
	}
	if err := helpers.EnsureBranchAccess(scope, branch.ID); err != nil {
		helpers.ErrorResponse(c, http.StatusForbidden, "Branch tidak dapat diakses", err.Error())
		return nil, nil, false
	}

	return scope, branch, true
}

func parseBusinessDate(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		location, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			location = time.FixedZone("Asia/Jakarta", 7*60*60)
		}
		return time.Now().In(location).Format("2006-01-02"), nil
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return "", err
	}
	return parsed.Format("2006-01-02"), nil
}

func storeStatusResponse(status *models.StoreStatus, summary models.StoreSalesSummary, branchID int64, businessDate string) gin.H {
	return gin.H{
		"status":                  status,
		"date":                    businessDate,
		"branch_id":               branchID,
		"posSales":                summary.POSSales,
		"posCashTotal":            summary.POSCashTotal,
		"posTransferTotal":        summary.POSTransferTotal,
		"expectedCash":            summary.ExpectedCash,
		"barbershopSales":         0,
		"barbershopCashSales":     0,
		"barbershopTransferSales": 0,
		"incomes":                 0,
		"expenses":                0,
		"reservationCashDeposits": 0,
		"reservationCashRefunds":  0,
		"storeBusinessCode":       nil,
	}
}
