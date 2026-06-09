package handler

import (
	"strings"

	"file_sys/backend/internal/dto"
	"file_sys/backend/internal/middleware"
	"file_sys/backend/internal/service"
	"file_sys/backend/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type OnlyOfficeHandler struct {
	ooService *service.OnlyOfficeService
}

func NewOnlyOfficeHandler(ooService *service.OnlyOfficeService) *OnlyOfficeHandler {
	return &OnlyOfficeHandler{ooService: ooService}
}

func (h *OnlyOfficeHandler) EditorConfig(c *gin.Context) {
	if h.ooService == nil {
		util.Error(c, 503, 50301, "OnlyOffice is not configured on this server")
		return
	}

	var req dto.EditorConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	userName := c.GetString("user_email")

	config, err := h.ooService.GenerateEditorConfig(c.Request.Context(), userID, userName, req.FileID, req.Mode)
	if err != nil {
		util.InternalError(c, "failed to generate editor config: "+err.Error())
		return
	}

	util.Success(c, config)
}

func extractOOToken(c *gin.Context) string {
	// Authorization header (Bearer <token>)
	auth := c.GetHeader("Authorization")
	if auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}
	// Query param (OnlyOffice DS uses ?token= when fetching files)
	if t := c.Query("token"); t != "" {
		return t
	}
	return ""
}

func verifyOOToken(token, secret string) bool {
	if token == "" {
		return false
	}
	_, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	return err == nil
}

func (h *OnlyOfficeHandler) Callback(c *gin.Context) {
	if h.ooService == nil {
		c.JSON(200, gin.H{"error": 0})
		return
	}

	var cb dto.OnlyOfficeCallback
	if err := c.ShouldBindJSON(&cb); err != nil {
		util.ValidationError(c, err.Error())
		return
	}

	// Get fileID from URL path (primary) or query token (legacy)
	fileID := c.Param("fileId")
	if fileID == "" {
		token := c.Query("token")
		claims, err := util.ValidateOOToken(token, h.ooService.GetJWTSecret())
		if err == nil && claims != nil {
			fileID = claims.FileID
		}
	}

	// Fallback: resolve file_id from document key via DB
	if fileID == "" {
		fileID = h.ooService.LookupFileIDByDocumentKey(c.Request.Context(), cb.Key)
	}

	if fileID == "" {
		c.JSON(200, gin.H{"error": 0})
		return
	}

	if err := h.ooService.HandleCallback(c.Request.Context(), &cb, fileID); err != nil {
		c.JSON(200, gin.H{"error": 1, "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"error": 0})
}

func (h *OnlyOfficeHandler) ServeFile(c *gin.Context) {
	if h.ooService == nil {
		util.Error(c, 503, 50301, "OnlyOffice is not configured")
		return
	}

	fileID := c.Param("fileId")

	// JWT verification: accept token from Authorization header or query param
	token := extractOOToken(c)
	if token != "" && !verifyOOToken(token, h.ooService.GetJWTSecret()) {
		util.Forbidden(c, "invalid token")
		return
	}

	reader, mimeType, size, err := h.ooService.GetFileStream(c.Request.Context(), fileID)
	if err != nil {
		util.NotFound(c, "file not found")
		return
	}
	defer reader.Close()

	c.Header("Content-Type", mimeType)
	c.Header("Content-Disposition", "inline")
	c.DataFromReader(200, size, mimeType, reader, nil)
}
