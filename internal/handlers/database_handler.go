package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type DatabaseHandler struct {
	DB *sql.DB
}

func NewDatabaseHandler(db *sql.DB) *DatabaseHandler {
	return &DatabaseHandler{
		DB: db,
	}
}

func (h *DatabaseHandler) Ping(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := h.DB.PingContext(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Database connection failed",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Database connection is healthy",
	})
}
