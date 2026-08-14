package apierror

import "github.com/gin-gonic/gin"

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

func Success(c *gin.Context, status int, data interface{}) {
	c.JSON(status, SuccessResponse{Success: true, Data: data})
}

func Fail(c *gin.Context, status int, code string, message string) {
	c.JSON(status, ErrorResponse{Success: false, Error: message, Code: code})
}

func BadRequest(c *gin.Context, message string) {
	Fail(c, 400, "BAD_REQUEST", message)
}

func Unauthorized(c *gin.Context, message string) {
	Fail(c, 401, "UNAUTHORIZED", message)
}

func NotFound(c *gin.Context, message string) {
	Fail(c, 404, "NOT_FOUND", message)
}

func Internal(c *gin.Context, message string) {
	Fail(c, 500, "INTERNAL_ERROR", message)
}
