package handler

import (
	"file_sys/backend/internal/repository"
	"file_sys/backend/internal/service"
	"file_sys/backend/internal/util"

	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	permRepo      *repository.PermissionRepo
	fileService    *service.FileService
	folderService  *service.FolderService
}

func NewPermissionHandler(permRepo *repository.PermissionRepo, fileService *service.FileService, folderService *service.FolderService) *PermissionHandler {
	return &PermissionHandler{permRepo: permRepo, fileService: fileService, folderService: folderService}
}

func (h *PermissionHandler) Update(c *gin.Context) {
	var req struct {
		Permission string `json:"permission" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}

	err := h.permRepo.Update(c.Request.Context(), c.Param("id"), req.Permission)
	if err != nil {
		util.NotFound(c, "permission not found")
		return
	}

	util.Success(c, nil)
}

func (h *PermissionHandler) Delete(c *gin.Context) {
	err := h.permRepo.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		util.NotFound(c, "permission not found")
		return
	}

	util.Success(c, nil)
}

func (h *PermissionHandler) SharedWithMeFolders(c *gin.Context) {
	userID, _ := c.Get("user_id")
	folders, err := h.permRepo.FindSharedFoldersByUser(c.Request.Context(), userID.(string))
	if err != nil {
		util.DatabaseError(c)
		return
	}

	util.Success(c, folders)
}

func (h *PermissionHandler) ListByFolder(c *gin.Context) {
	perms, err := h.permRepo.FindByResource(c.Request.Context(), "folder", c.Param("id"))
	if err != nil {
		util.DatabaseError(c)
		return
	}

	util.Success(c, perms)
}

func (h *PermissionHandler) SharedWithMeFiles(c *gin.Context) {
	userID, _ := c.Get("user_id")
	files, err := h.permRepo.FindSharedFilesByUser(c.Request.Context(), userID.(string))
	if err != nil {
		util.DatabaseError(c)
		return
	}

	util.Success(c, files)
}

func (h *PermissionHandler) ListByFile(c *gin.Context) {
	perms, err := h.permRepo.FindByResource(c.Request.Context(), "file", c.Param("id"))
	if err != nil {
		util.DatabaseError(c)
		return
	}

	util.Success(c, perms)
}
