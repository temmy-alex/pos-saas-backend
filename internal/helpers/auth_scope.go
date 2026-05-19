package helpers

import (
	"errors"

	"pos-saas-backend/internal/auth"

	"github.com/gin-gonic/gin"
)

type AuthScope struct {
	UserID   int64
	Role     string
	StoreID  int64
	BranchID int64
}

func GetAuthScope(c *gin.Context) (*AuthScope, error) {
	authUserValue, exists := c.Get("auth_user")
	if !exists {
		return nil, errors.New("auth user not found")
	}

	claims, ok := authUserValue.(*auth.Claims)
	if !ok {
		return nil, errors.New("invalid auth user claims")
	}

	scope := &AuthScope{
		UserID: claims.UserID,
		Role:   claims.Role,
	}

	if claims.StoreID != nil {
		scope.StoreID = *claims.StoreID
	}

	if claims.BranchID != nil {
		scope.BranchID = *claims.BranchID
	}

	return scope, nil
}

func ApplyStoreScope(scope *AuthScope, requestedStoreID int64) (int64, error) {
	if scope.Role == "superadmin" {
		return requestedStoreID, nil
	}

	if scope.StoreID <= 0 {
		return 0, errors.New("user store scope is not configured")
	}

	if requestedStoreID > 0 && requestedStoreID != scope.StoreID {
		return 0, errors.New("you do not have access to this store")
	}

	return scope.StoreID, nil
}

func ApplyBranchScope(scope *AuthScope, requestedBranchID int64) (int64, error) {
	if scope.Role == "superadmin" || scope.Role == "admin" {
		return requestedBranchID, nil
	}

	if scope.Role == "cashier" {
		if scope.BranchID <= 0 {
			return 0, errors.New("user branch scope is not configured")
		}

		if requestedBranchID > 0 && requestedBranchID != scope.BranchID {
			return 0, errors.New("you do not have access to this branch")
		}

		return scope.BranchID, nil
	}

	return requestedBranchID, nil
}

func EnsureStoreAccess(scope *AuthScope, storeID int64) error {
	if scope.Role == "superadmin" {
		return nil
	}

	if scope.StoreID <= 0 {
		return errors.New("user store scope is not configured")
	}

	if storeID != scope.StoreID {
		return errors.New("you do not have access to this store")
	}

	return nil
}

func EnsureBranchAccess(scope *AuthScope, branchID int64) error {
	if scope.Role == "superadmin" || scope.Role == "admin" {
		return nil
	}

	if scope.Role == "cashier" {
		if scope.BranchID <= 0 {
			return errors.New("user branch scope is not configured")
		}

		if branchID != scope.BranchID {
			return errors.New("you do not have access to this branch")
		}
	}

	return nil
}
