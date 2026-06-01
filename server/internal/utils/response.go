package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SuccessResponse is the shape of every successful JSON response.
// Consistent structure makes the frontend's life easier.
type SuccessResponse struct {
	Data any `json:"data"`
}

// ErrorResponse is the shape of every error JSON response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// OK sends a 200 response with data wrapped in { "data": ... }
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, SuccessResponse{Data: data})
}

// BadRequest sends a 400 with an error message.
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{Error: message})
}

// Unauthorized sends a 401.
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, ErrorResponse{Error: message})
}

// Forbidden sends a 403.
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, ErrorResponse{Error: message})
}

// NotFound sends a 404.
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, ErrorResponse{Error: message})
}

// InternalError sends a 500.
func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{Error: message})
}
