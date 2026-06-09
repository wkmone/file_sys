package middleware

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

// JSONRecovery returns JSON on panics instead of the default HTML page.
func JSONRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Check for broken pipe — don't log, just abort
				if ne, ok := r.(*net.OpError); ok && ne.Op == "write" {
					c.Abort()
					return
				}

				stack := string(debug.Stack())

				// Log full details to server console
				log.Printf("[PANIC] %s %s | %v\n%s", c.Request.Method, c.Request.URL.Path, r, stack)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    50000,
					"message": fmt.Sprintf("服务器内部错误: %v", r),
				})
			}
		}()
		c.Next()
	}
}

// ErrorLogger logs request/response details for debugging.
// Must be used after c.Next() to capture the response status.
func ErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Capture request body for logging on error
		var reqBody string
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			reqBody = string(bodyBytes)
			// Restore body for downstream handlers
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		status := c.Writer.Status()
		if status >= 400 {
			log.Printf("[ERROR] %s %s | %d | %s | body: %s",
				c.Request.Method, c.Request.URL.Path,
				status, time.Since(start).Round(time.Millisecond),
				truncate(reqBody, 500),
			)
		}
	}
}

// NoRouteJSON returns JSON for 404 instead of the default text response.
func NoRouteJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": fmt.Sprintf("接口不存在: %s %s", c.Request.Method, c.Request.URL.Path),
		})
	}
}

// NoMethodJSON returns JSON for 405 instead of the default text response.
func NoMethodJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{
			"code":    40500,
			"message": fmt.Sprintf("方法不允许: %s", c.Request.Method),
		})
	}
}

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

// SuppressGoLogo suppresses the Go logo ASCII art on startup from Gin's debug print.
func SuppressGoLogo() {
	// Gin writes its banner to os.Stdout on init; we can't suppress it directly,
	// but we can silence the "debug print" routes output in release mode.
	if os.Getenv("GIN_MODE") == "" && os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
}
