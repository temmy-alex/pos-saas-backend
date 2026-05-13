package middlewares

import (
	"net/http"

	"pos-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type RoleMiddleware struct{}

func NewRoleMiddleware() *RoleMiddleware {
	return &RoleMiddleware{}
}

func (m *RoleMiddleware) RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authUserValue, exists := c.Get("auth_user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Unauthorized",
				"error":   "auth user not found",
			})
			return
		}

		claims, ok := authUserValue.(*services.AuthClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Unauthorized",
				"error":   "invalid auth user claims",
			})
			return
		}

		for _, role := range allowedRoles {
			if claims.Role == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "Forbidden",
			"error":   "you do not have permission to access this resource",
		})
	}
}
