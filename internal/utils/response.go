package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RespondSuccess sends a standardized success JSON response.
func RespondSuccess(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, gin.H{
		"status":  status,
		"message": message,
		"data":    data,
	})
}

// RespondError sends a standardized error JSON response.
func RespondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"status":  status,
		"message": message,
	})
}

// RespondValidationError sends a 400 response with validation error details.
func RespondValidationError(c *gin.Context, message string, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status":  http.StatusBadRequest,
		"message": message,
		"error":   err.Error(),
	})
}
