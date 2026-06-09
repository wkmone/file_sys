package handler

import (
	"file_sys/backend/internal/middleware"
	"file_sys/backend/internal/model"
	"file_sys/backend/internal/service"
	"file_sys/backend/internal/util"

	"github.com/gin-gonic/gin"
)

type TrashHandler struct {
	fileService   *service.FileService
	folderService *service.FolderService
}

func NewTrashHandler(fileService *service.FileService, folderService *service.FolderService) *TrashHandler {
	return &TrashHandler{fileService: fileService, folderService: folderService}
}

type trashItem struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	DeletedAt string  `json:"deleted_at"`
	FileSize  int64   `json:"file_size,omitempty"`
	FileExt   string  `json:"file_ext,omitempty"`
}

func (h *TrashHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	teamID := c.Query("team_id")

	var files []model.File
	var folders []model.Folder

	if teamID != "" {
		files, _ = h.fileService.FindDeletedByTeam(c.Request.Context(), teamID)
		folders, _ = h.folderService.FindDeletedByTeam(c.Request.Context(), teamID)
	} else {
		files, _ = h.fileService.FindDeleted(c.Request.Context(), userID)
		folders, _ = h.folderService.FindDeleted(c.Request.Context(), userID)
	}

	items := make([]trashItem, 0, len(files)+len(folders))
	for _, f := range files {
		deletedAt := ""
		if f.DeletedAt != nil {
			deletedAt = f.DeletedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		items = append(items, trashItem{
			ID: f.ID, Name: f.OriginalName, Type: "file",
			DeletedAt: deletedAt,
			FileSize:  f.FileSize, FileExt: f.FileExt,
		})
	}
	for _, f := range folders {
		deletedAt := ""
		if f.DeletedAt != nil {
			deletedAt = f.DeletedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		items = append(items, trashItem{
			ID: f.ID, Name: f.Name, Type: "folder",
			DeletedAt: deletedAt,
		})
	}
	util.Success(c, gin.H{"items": items})
}

func (h *TrashHandler) Restore(c *gin.Context) {
	typ := c.Param("type")
	id := c.Param("id")

	var err error
	if typ == "file" {
		err = h.fileService.Restore(c.Request.Context(), id)
	} else {
		err = h.folderService.Restore(c.Request.Context(), id)
	}
	if err != nil {
		util.InternalError(c, "restore failed")
		return
	}
	util.Success(c, nil)
}

func (h *TrashHandler) PermanentDelete(c *gin.Context) {
	typ := c.Param("type")
	id := c.Param("id")

	var err error
	if typ == "file" {
		err = h.fileService.PermanentDelete(c.Request.Context(), id)
	} else {
		err = h.folderService.PermanentDelete(c.Request.Context(), id)
	}
	if err != nil {
		util.InternalError(c, "delete failed")
		return
	}
	util.Success(c, nil)
}
