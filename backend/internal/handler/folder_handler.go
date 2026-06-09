package handler

import (
	"strconv"

	"file_sys/backend/internal/dto"
	"file_sys/backend/internal/middleware"
	"file_sys/backend/internal/service"
	"file_sys/backend/internal/util"

	"github.com/gin-gonic/gin"
)

type FolderHandler struct {
	folderService *service.FolderService
}

func NewFolderHandler(folderService *service.FolderService) *FolderHandler {
	return &FolderHandler{folderService: folderService}
}

func (h *FolderHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	parentID := c.Query("parent_id")
	teamID := c.Query("team_id")

	var pid *string
	if parentID != "" {
		pid = &parentID
	}

	var folders interface{}
	var total int64
	var err error

	if teamID != "" {
		folders, total, err = h.folderService.ListByTeam(c.Request.Context(), teamID, pid, page, pageSize)
	} else {
		folders, total, err = h.folderService.ListByParent(c.Request.Context(), pid, middleware.GetUserID(c), page, pageSize)
	}

	if err != nil {
		util.InternalError(c, "failed to list folders")
		return
	}

	util.Success(c, dto.PaginatedResponse{
		Items:      folders,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	})
}

func (h *FolderHandler) Create(c *gin.Context) {
	var req dto.CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}

	folder, err := h.folderService.Create(c.Request.Context(), &req, middleware.GetUserID(c))
	if err != nil {
		util.InternalError(c, "create folder failed")
		return
	}
	util.Created(c, folder)
}

func (h *FolderHandler) Get(c *gin.Context) {
	f, err := h.folderService.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		util.NotFound(c, "folder not found")
		return
	}
	util.Success(c, f)
}

func (h *FolderHandler) Update(c *gin.Context) {
	var req dto.UpdateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}
	if err := h.folderService.Update(c.Request.Context(), c.Param("id"), &req); err != nil {
		util.InternalError(c, "update failed")
		return
	}
	util.Success(c, nil)
}

func (h *FolderHandler) Delete(c *gin.Context) {
	if err := h.folderService.Delete(c.Request.Context(), c.Param("id")); err != nil {
		util.InternalError(c, "delete failed")
		return
	}
	util.Success(c, nil)
}

func (h *FolderHandler) Tree(c *gin.Context) {
	folders, err := h.folderService.GetAllByOwner(c.Request.Context(), middleware.GetUserID(c))
	if err != nil {
		util.InternalError(c, "failed to get folder tree")
		return
	}
	util.Success(c, folders)
}

func (h *FolderHandler) Share(c *gin.Context) {
	var req dto.ShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}
	if err := h.folderService.Share(c.Request.Context(), c.Param("id"), middleware.GetUserID(c), &req); err != nil {
		util.InternalError(c, "share failed")
		return
	}
	util.Success(c, nil)
}
