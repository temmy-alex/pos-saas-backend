package handlers

import (
	"context"
	"net/http"
	"time"

	"pos-saas-backend/internal/auth"
	"pos-saas-backend/internal/helpers"
	"pos-saas-backend/internal/requests"
	"pos-saas-backend/internal/responses"
	"pos-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	AuthService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		AuthService: authService,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request requests.LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	loginResponse, err := h.AuthService.Login(ctx, request)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Login failed", err.Error())
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Login successful", loginResponse)
}

func (h *AuthHandler) Me(c *gin.Context) {
	authUserValue, exists := c.Get("auth_user")
	if !exists {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "auth user not found")
		return
	}

	claims, ok := authUserValue.(*auth.Claims)
	if !ok {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "invalid auth user claims")
		return
	}

	userResponse := responses.AuthUserResponse{
		ID:       claims.UserID,
		StoreID:  claims.StoreID,
		BranchID: claims.BranchID,
		Name:     claims.Name,
		Email:    claims.Email,
		Role:     claims.Role,
	}

	helpers.SuccessResponse(c, http.StatusOK, "Authenticated user retrieved successfully", userResponse)
}
