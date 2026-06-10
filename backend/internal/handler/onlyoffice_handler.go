package handler

import (
	"io"
	"log"

	"github.com/gin-gonic/gin"

	"file_sys/backend/internal/dto"
	"file_sys/backend/internal/service"
	"file_sys/backend/internal/util"
)

type OnlyOfficeHandler struct {
	ooService *service.OnlyOfficeService
}

func NewOnlyOfficeHandler(ooService *service.OnlyOfficeService) *OnlyOfficeHandler {
	return &OnlyOfficeHandler{ooService: ooService}
}

func (h *OnlyOfficeHandler) EditorConfig(c *gin.Context) {
	var req dto.EditorConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	resp, err := h.ooService.GenerateEditorConfig(c.Request.Context(), userID.(string), "", req.FileID, req.Mode)
	if err != nil {
		util.FileNotFound(c)
		return
	}

	util.Success(c, resp)
}

func (h *OnlyOfficeHandler) ServeFile(c *gin.Context) {
	fileID := c.Param("fileId")
	reader, mimeType, _, err := h.ooService.GetFileStream(c.Request.Context(), fileID)
	if err != nil {
		log.Printf("ServeFile error for fileID=%q: %v", fileID, err)
		util.FileNotFound(c)
		return
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		util.InternalError(c, "failed to read file")
		return
	}

	c.Header("Content-Type", mimeType)
	c.Data(200, mimeType, content)
}

func (h *OnlyOfficeHandler) Callback(c *gin.Context) {
	fileID := c.Param("fileId")

	var cb dto.OnlyOfficeCallback
	if err := c.ShouldBindJSON(&cb); err != nil {
		util.ValidationError(c, "invalid callback payload")
		return
	}

	if err := h.ooService.HandleCallback(c.Request.Context(), &cb, fileID); err != nil {
		util.InternalError(c, "failed to process callback")
		return
	}

	util.Success(c, nil)
}
