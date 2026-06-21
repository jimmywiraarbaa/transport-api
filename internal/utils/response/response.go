// Package response provides a consistent JSON envelope used by every HTTP
// handler across feature packages.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope is the canonical API response shape.
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

// Meta carries pagination/extra info alongside data.
type Meta struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

// ErrorBody describes a failure.
type ErrorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []FieldError  `json:"details,omitempty"`
}

// FieldError is a single validation field error.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// OK writes a 200 with success envelope.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data})
}

// Created writes a 201 with success envelope.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Envelope{Success: true, Data: data})
}

// OKWithMeta writes a 200 with data and pagination meta.
func OKWithMeta(c *gin.Context, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data, Meta: meta})
}

// NoContent writes a 204.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error writes an error envelope with the given status and code.
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, Envelope{
		Success: false,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

// ValidationError writes a 422 with field-level details.
func ValidationError(c *gin.Context, details []FieldError) {
	c.JSON(http.StatusUnprocessableEntity, Envelope{
		Success: false,
		Error: &ErrorBody{
			Code:    "VALIDATION_ERROR",
			Message: "The given data was invalid",
			Details: details,
		},
	})
}

// Standard error code → status helpers, so handlers stay terse.
func NotFound(c *gin.Context, msg string)    { Error(c, http.StatusNotFound, "NOT_FOUND", or(msg, "Resource not found")) }
func Unauthorized(c *gin.Context, msg string) { Error(c, http.StatusUnauthorized, "UNAUTHORIZED", or(msg, "Authentication required")) }
func Forbidden(c *gin.Context, msg string)   { Error(c, http.StatusForbidden, "FORBIDDEN", or(msg, "Access denied")) }
func Conflict(c *gin.Context, msg string)    { Error(c, http.StatusConflict, "CONFLICT", or(msg, "Resource already exists")) }
func Internal(c *gin.Context, err error)     { Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong") }

func or(msg, fallback string) string {
	if msg != "" {
		return msg
	}
	return fallback
}
