package handler

import (
	"encoding/json"
	"mime/multipart"
	"strconv"

	"file_sys/backend/internal/dto"
	"file_sys/backend/internal/middleware"
	"file_sys/backend/internal/model"
	"file_sys/backend/internal/service"
	"file_sys/backend/internal/util"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	fileService *service.FileService
}

func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

func (h *FileHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	folderID := c.Query("folder_id")
	teamID := c.Query("team_id")

	var fid *string
	if folderID != "" {
		fid = &folderID
	}

	var files []model.File
	var total int64
	var err error

	if teamID != "" {
		files, total, err = h.fileService.ListByTeam(c.Request.Context(), teamID, fid, page, pageSize)
	} else {
		files, total, err = h.fileService.ListByFolder(c.Request.Context(), fid, middleware.GetUserID(c), page, pageSize)
	}

	if err != nil {
		util.InternalError(c, "failed to list files")
		return
	}

	util.Success(c, dto.PaginatedResponse{
		Items:      files,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	})
}

func (h *FileHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		util.ValidationError(c, "missing file")
		return
	}
	defer file.Close()

	folderID := c.PostForm("folder_id")
	rawTeamID := c.PostForm("team_id")
	var teamID *string
	if rawTeamID != "" {
		teamID = &rawTeamID
	}
	resp, err := h.fileService.Upload(c.Request.Context(), file, header.Filename, folderID, middleware.GetUserID(c), teamID)
	if err != nil {
		util.InternalError(c, "upload failed: "+err.Error())
		return
	}
	util.Created(c, resp)
}

func (h *FileHandler) BatchUpload(c *gin.Context) {
	// Parse multipart form with 100MB limit
	if err := c.Request.ParseMultipartForm(100 << 20); err != nil {
		util.ValidationError(c, "request too large")
		return
	}

	manifestStr := c.PostForm("manifest")
	if manifestStr == "" {
		util.ValidationError(c, "missing manifest")
		return
	}

	var manifest []dto.BatchUploadManifestEntry
	if err := json.Unmarshal([]byte(manifestStr), &manifest); err != nil {
		util.ValidationError(c, "invalid manifest JSON: "+err.Error())
		return
	}

	var fileHeaders []*multipart.FileHeader
	if c.Request.MultipartForm != nil && c.Request.MultipartForm.File != nil {
		fileHeaders = c.Request.MultipartForm.File["files"]
	}

	folderID := c.PostForm("folder_id")
	rawTeamID := c.PostForm("team_id")
	var teamID *string
	if rawTeamID != "" {
		teamID = &rawTeamID
	}

	resp, err := h.fileService.BatchUpload(c.Request.Context(), fileHeaders, manifestStr, folderID, middleware.GetUserID(c), teamID)
	if err != nil {
		util.InternalError(c, "batch upload failed: "+err.Error())
		return
	}
	util.Created(c, resp)
}

func (h *FileHandler) Get(c *gin.Context) {
	f, err := h.fileService.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		util.NotFound(c, "file not found")
		return
	}
	util.Success(c, f)
}

func (h *FileHandler) Update(c *gin.Context) {
	var req dto.UpdateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}
	if err := h.fileService.Update(c.Request.Context(), c.Param("id"), &req); err != nil {
		util.InternalError(c, "update failed")
		return
	}
	util.Success(c, nil)
}

func (h *FileHandler) Delete(c *gin.Context) {
	if err := h.fileService.Delete(c.Request.Context(), c.Param("id")); err != nil {
		util.InternalError(c, "delete failed")
		return
	}
	util.Success(c, nil)
}

func (h *FileHandler) Download(c *gin.Context) {
	key, err := h.fileService.GetStorageKey(c.Request.Context(), c.Param("id"))
	if err != nil {
		util.NotFound(c, "file not found")
		return
	}

	reader, info, err := h.fileService.RetrieveStorage(c.Request.Context(), key)
	if err != nil {
		util.InternalError(c, "file retrieval failed")
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", "inline")
	c.DataFromReader(200, info.Size, info.ContentType, reader, nil)
}

func (h *FileHandler) Copy(c *gin.Context) {
	// TODO: implement copy
	util.Success(c, nil)
}

func (h *FileHandler) Share(c *gin.Context) {
	var req dto.ShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}
	if err := h.fileService.Share(c.Request.Context(), c.Param("id"), middleware.GetUserID(c), &req); err != nil {
		util.InternalError(c, "share failed")
		return
	}
	util.Success(c, nil)
}

