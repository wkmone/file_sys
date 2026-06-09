package middleware

import (
	"strings"

	"file_sys/backend/internal/util"

	"github.com/gin-gonic/gin"
)

func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				util.Unauthorized(c, "invalid authorization format")
				return
			}
			tokenStr = parts[1]
		} else {
			// Fallback: support ?token= query parameter for browser-opened links (PDF preview, download)
			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			util.Unauthorized(c, "missing authorization header")
			return
		}

		claims, err := util.ValidateAccessToken(tokenStr, jwtSecret)
		if err != nil {
			util.Unauthorized(c, "invalid or expired token")
			return
		}

		c.Set("user_id", claims.Sub)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("user_role")
		if role != "admin" && role != "super_admin" {
			util.Forbidden(c, "admin access required")
			return
		}
		c.Next()
	}
}

func GetUserID(c *gin.Context) string {
	id, _ := c.Get("user_id")
	return id.(string)
}
