package util

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Code: 0, Message: "created", Data: data})
}

func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.AbortWithStatusJSON(httpStatus, Response{Code: code, Message: message})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, 40000, message)
}

func ValidationError(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, 40001, message)
}

func InvalidEmail(c *gin.Context) {
	Error(c, http.StatusBadRequest, 40002, "invalid email format")
}

func InvalidPassword(c *gin.Context) {
	Error(c, http.StatusBadRequest, 40003, "password must be at least 8 characters with letters and numbers")
}

func InvalidFile(c *gin.Context) {
	Error(c, http.StatusBadRequest, 40004, "invalid file")
}

func FileTooLarge(c *gin.Context, maxSize int64) {
	Error(c, http.StatusBadRequest, 40005, fmt.Sprintf("file size exceeds %d bytes", maxSize))
}

func InvalidFileType(c *gin.Context) {
	Error(c, http.StatusBadRequest, 40006, "unsupported file type")
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, 40100, message)
}

func InvalidToken(c *gin.Context) {
	Error(c, http.StatusUnauthorized, 40101, "invalid or expired token")
}

func ExpiredToken(c *gin.Context) {
	Error(c, http.StatusUnauthorized, 40102, "token has expired")
}

func MissingToken(c *gin.Context) {
	Error(c, http.StatusUnauthorized, 40103, "missing authorization token")
}

func InvalidCredentials(c *gin.Context) {
	Error(c, http.StatusUnauthorized, 40104, "invalid email or password")
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, 40300, message)
}

func InsufficientPermissions(c *gin.Context) {
	Error(c, http.StatusForbidden, 40301, "insufficient permissions")
}

func AdminRequired(c *gin.Context) {
	Error(c, http.StatusForbidden, 40302, "admin privileges required")
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, 40400, message)
}

func UserNotFound(c *gin.Context) {
	Error(c, http.StatusNotFound, 40401, "user not found")
}

func FileNotFound(c *gin.Context) {
	Error(c, http.StatusNotFound, 40402, "file not found")
}

func FolderNotFound(c *gin.Context) {
	Error(c, http.StatusNotFound, 40403, "folder not found")
}

func TeamNotFound(c *gin.Context) {
	Error(c, http.StatusNotFound, 40404, "team not found")
}

func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, 40900, message)
}

func EmailExists(c *gin.Context) {
	Error(c, http.StatusConflict, 40901, "email already registered")
}

func FileExists(c *gin.Context) {
	Error(c, http.StatusConflict, 40902, "file already exists")
}

func FolderExists(c *gin.Context) {
	Error(c, http.StatusConflict, 40903, "folder already exists")
}

func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, 50000, message)
}

func DatabaseError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, 50001, "database operation failed")
}

func StorageError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, 50002, "storage operation failed")
}

func ExternalServiceError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, 50003, "external service error")
}