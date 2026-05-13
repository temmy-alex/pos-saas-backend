package helpers

import (
	"github.com/gin-gonic/gin"
)

func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, gin.H{
		"status":  "success",
		"message": message,
		"data":    data,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, message string, err string) {
	c.JSON(statusCode, gin.H{
		"status":  "error",
		"message": message,
		"error":   err,
	})
}
