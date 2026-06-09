package handler

import (
	"file_sys/backend/internal/service"
	"file_sys/backend/internal/util"

	"github.com/gin-gonic/gin"
)

type VersionHandler struct {
	fileService *service.FileService
}

func NewVersionHandler(fileService *service.FileService) *VersionHandler {
	return &VersionHandler{fileService: fileService}
}

func (h *VersionHandler) List(c *gin.Context) {
	versions, err := h.fileService.GetVersions(c.Request.Context(), c.Param("id"))
	if err != nil {
		util.InternalError(c, "failed to list versions")
		return
	}
	util.Success(c, versions)
}

func (h *VersionHandler) Get(c *gin.Context) {
	v, err := h.fileService.GetVersion(c.Request.Context(), c.Param("vid"))
	if err != nil {
		util.NotFound(c, "version not found")
		return
	}
	util.Success(c, v)
}

func (h *VersionHandler) Download(c *gin.Context) {
	reader, info, err := h.fileService.DownloadVersion(c.Request.Context(), c.Param("vid"))
	if err != nil {
		util.NotFound(c, "version not found")
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", "inline")
	c.DataFromReader(200, info.Size, info.ContentType, reader, nil)
}

func (h *VersionHandler) Restore(c *gin.Context) {
	if err := h.fileService.RestoreVersion(c.Request.Context(), c.Param("id"), c.Param("vid")); err != nil {
		util.InternalError(c, "restore failed")
		return
	}
	util.Success(c, nil)
}
