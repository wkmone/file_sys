package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"file_sys/backend/internal/dto"
	"file_sys/backend/internal/middleware"
	"file_sys/backend/internal/model"
	"file_sys/backend/internal/service"
	"file_sys/backend/internal/util"
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

	userID := middleware.GetUserID(c)

	var parentIDPtr *string
	if parentID != "" {
		parentIDPtr = &parentID
	}

	var folders []model.Folder
	var total int64
	var err error

	if teamID != "" {
		folders, total, err = h.folderService.ListByTeam(c.Request.Context(), teamID, parentIDPtr, userID, page, pageSize)
	} else {
		folders, total, err = h.folderService.ListByParent(c.Request.Context(), parentIDPtr, userID, page, pageSize)
	}

	if err != nil {
		util.DatabaseError(c)
		return
	}

	// Convert to interface slice for pagination
	items := make([]interface{}, len(folders))
	for i, f := range folders {
		items[i] = f
	}

	util.Success(c, dto.PaginatedResponse{
		Items:      items,
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

	userID := middleware.GetUserID(c)
	folder, err := h.folderService.Create(c.Request.Context(), &req, userID)
	if err != nil {
		util.FolderExists(c)
		return
	}

	util.Created(c, folder)
}

func (h *FolderHandler) Get(c *gin.Context) {
	folder, err := h.folderService.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		util.FolderNotFound(c)
		return
	}

	util.Success(c, folder)
}

func (h *FolderHandler) Update(c *gin.Context) {
	var req dto.UpdateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}

	err := h.folderService.Update(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		util.FolderNotFound(c)
		return
	}

	util.Success(c, nil)
}

func (h *FolderHandler) Delete(c *gin.Context) {
	err := h.folderService.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		util.FolderNotFound(c)
		return
	}

	util.Success(c, nil)
}

func (h *FolderHandler) Tree(c *gin.Context) {
	userID := middleware.GetUserID(c)
	folders, err := h.folderService.GetAllByOwner(c.Request.Context(), userID)
	if err != nil {
		util.DatabaseError(c)
		return
	}

	// Build tree - simple flat list for now
	util.Success(c, gin.H{
		"folders": folders,
	})
}

func (h *FolderHandler) Share(c *gin.Context) {
	var req dto.ShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	err := h.folderService.Share(c.Request.Context(), c.Param("id"), userID, &req)
	if err != nil {
		util.InternalError(c, "share failed")
		return
	}

	util.Success(c, nil)
}
