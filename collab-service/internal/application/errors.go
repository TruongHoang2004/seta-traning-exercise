package application

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HTTPError là custom error cho phép kèm HTTP status code
type HTTPError struct {
	Code    int    // HTTP status code, ví dụ 403, 404
	Message string // Thông báo lỗi hiển thị cho client
}

// Error implement interface error
func (e *HTTPError) Error() string {
	return e.Message
}

// Helper functions tạo các lỗi phổ biến

// NewBadRequestError trả về lỗi 400
func NewBadRequestError(msg string) *HTTPError {
	return &HTTPError{
		Code:    http.StatusBadRequest,
		Message: msg,
	}
}

// NewUnauthorizedError trả về lỗi 401
func NewUnauthorizedError(msg string) *HTTPError {
	return &HTTPError{
		Code:    http.StatusUnauthorized,
		Message: msg,
	}
}

// NewForbiddenError trả về lỗi 403
func NewForbiddenError(msg string) *HTTPError {
	return &HTTPError{
		Code:    http.StatusForbidden,
		Message: msg,
	}
}

// NewNotFoundError trả về lỗi 404
func NewNotFoundError(msg string) *HTTPError {
	return &HTTPError{
		Code:    http.StatusNotFound,
		Message: msg,
	}
}

// NewConflictError trả về lỗi 409
func NewConflictError(msg string) *HTTPError {
	return &HTTPError{
		Code:    http.StatusConflict,
		Message: msg,
	}
}

// NewInternalServerError trả về lỗi 500
func NewInternalServerError(msg string) *HTTPError {
	return &HTTPError{
		Code:    http.StatusInternalServerError,
		Message: msg,
	}
}

// HandleError kiểm tra error và trả response JSON tương ứng
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	if httpErr, ok := err.(*HTTPError); ok {
		c.JSON(httpErr.Code, gin.H{"error": httpErr.Message})
		return
	}

	// Default: Internal Server Error
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
