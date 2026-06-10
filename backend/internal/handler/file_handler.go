package handler

import (
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"file_sys/backend/internal/dto"
	"file_sys/backend/internal/middleware"
	"file_sys/backend/internal/model"
	"file_sys/backend/internal/repository"
	"file_sys/backend/internal/service"
	"file_sys/backend/internal/util"
)

type FileHandler struct {
	fileService *service.FileService
	permRepo    *repository.PermissionRepo
	maxFileSize int64
}

func NewFileHandler(fileService *service.FileService, permRepo *repository.PermissionRepo) *FileHandler {
	return &FileHandler{fileService: fileService, permRepo: permRepo, maxFileSize: 100 << 20}
}

func (h *FileHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	folderID := c.Query("folder_id")
	teamID := c.Query("team_id")

	userID := middleware.GetUserID(c)

	var files []model.File
	var total int64
	var err error

	var folderIDPtr *string
	if folderID != "" {
		folderIDPtr = &folderID
	}

	if teamID != "" {
		files, total, err = h.fileService.ListByTeam(c.Request.Context(), teamID, folderIDPtr, userID, page, pageSize)
	} else {
		files, total, err = h.fileService.ListByFolder(c.Request.Context(), folderIDPtr, userID, page, pageSize)
	}

	if err != nil {
		log.Printf("FileHandler.List error: %v", err)
		util.DatabaseError(c)
		return
	}

	// Convert to interface slice for pagination
	items := make([]interface{}, len(files))
	for i, f := range files {
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

func (h *FileHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		util.InvalidFile(c)
		return
	}

	if file.Size > h.maxFileSize {
		util.FileTooLarge(c, h.maxFileSize)
		return
	}

	f, err := file.Open()
	if err != nil {
		util.InvalidFile(c)
		return
	}
	defer f.Close()

	folderID := c.PostForm("folder_id")
	teamID := c.PostForm("team_id")
	userID := middleware.GetUserID(c)

	var teamIDPtr *string
	if teamID != "" {
		teamIDPtr = &teamID
	}

	resp, err := h.fileService.Upload(c.Request.Context(), f, file.Filename, folderID, userID, teamIDPtr)
	if err != nil {
		if strings.Contains(err.Error(), "storage") {
			util.StorageError(c)
		} else {
			util.InternalError(c, "upload failed")
		}
		return
	}

	util.Created(c, resp)
}

func (h *FileHandler) BatchUpload(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(h.maxFileSize * 5); err != nil {
		util.BadRequest(c, "request too large")
		return
	}

	manifestJSON := c.PostForm("manifest")
	folderID := c.PostForm("folder_id")
	teamID := c.PostForm("team_id")
	userID := middleware.GetUserID(c)

	var teamIDPtr *string
	if teamID != "" {
		teamIDPtr = &teamID
	}

	// Parse manifest to determine file ordering
	var manifest []dto.BatchUploadManifestEntry
	if err := json.Unmarshal([]byte(manifestJSON), &manifest); err != nil {
		util.BadRequest(c, "invalid manifest JSON")
		return
	}

	// Build filename → fileHeader map
	nameToFile := make(map[string]*multipart.FileHeader)
	for _, fhs := range c.Request.MultipartForm.File {
		for _, fh := range fhs {
			nameToFile[fh.Filename] = fh
		}
	}

	// Order files by manifest entry path
	var orderedFiles []*multipart.FileHeader
	for _, entry := range manifest {
		if entry.IsDirectory {
			continue
		}
		base := filepath.Base(entry.Path)
		if fh, ok := nameToFile[base]; ok {
			orderedFiles = append(orderedFiles, fh)
		}
	}

	resp, err := h.fileService.BatchUpload(c.Request.Context(), orderedFiles, manifestJSON, folderID, userID, teamIDPtr)
	if err != nil {
		util.InternalError(c, "batch upload failed")
		return
	}

	util.Created(c, resp)
}

func (h *FileHandler) CreateBlank(c *gin.Context) {
	var req dto.CreateBlankFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	teamID := c.Query("team_id")
	var teamIDPtr *string
	if teamID != "" {
		teamIDPtr = &teamID
	}

	file, err := h.fileService.CreateBlank(c.Request.Context(), req.Name, req.FileExt, req.FolderID, userID, teamIDPtr)
	if err != nil {
		util.InternalError(c, "failed to create blank file")
		return
	}

	util.Created(c, file)
}

func (h *FileHandler) Get(c *gin.Context) {
	userID := middleware.GetUserID(c)
	file, err := h.fileService.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		util.FileNotFound(c)
		return
	}

	// Owner always has access
	if file.OwnerID == userID {
		util.Success(c, file)
		return
	}

	// Check explicit permission
	if h.permRepo != nil {
		if _, err := h.permRepo.FindByUserAndFile(c.Request.Context(), userID, file.ID); err == nil {
			util.Success(c, file)
			return
		}
	}

	util.InsufficientPermissions(c)
}

func (h *FileHandler) Update(c *gin.Context) {
	var req dto.UpdateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}

	err := h.fileService.Update(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		util.FileNotFound(c)
		return
	}

	util.Success(c, nil)
}

func (h *FileHandler) Delete(c *gin.Context) {
	err := h.fileService.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		util.FileNotFound(c)
		return
	}

	util.Success(c, nil)
}

func (h *FileHandler) Download(c *gin.Context) {
	fileID := c.Param("id")
	file, err := h.fileService.GetByID(c.Request.Context(), fileID)
	if err != nil {
		util.FileNotFound(c)
		return
	}

	key, err := h.fileService.GetStorageKey(c.Request.Context(), fileID)
	if err != nil {
		util.FileNotFound(c)
		return
	}

	reader, info, err := h.fileService.RetrieveStorage(c.Request.Context(), key)
	if err != nil {
		util.StorageError(c)
		return
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		util.InternalError(c, "failed to read file")
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\""+file.Name+"\"")
	c.Data(200, info.ContentType, content)
}

func (h *FileHandler) Copy(c *gin.Context) {
	// Simple copy - read original and re-upload as new file
	file, err := h.fileService.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		util.FileNotFound(c)
		return
	}

	key, err := h.fileService.GetStorageKey(c.Request.Context(), file.ID)
	if err != nil {
		util.FileNotFound(c)
		return
	}

	reader, _, err := h.fileService.RetrieveStorage(c.Request.Context(), key)
	if err != nil {
		util.StorageError(c)
		return
	}
	defer reader.Close()

	userID := middleware.GetUserID(c)
	newName := file.Name + " (copy)"

	folderID := ""
	if file.FolderID != nil {
		folderID = *file.FolderID
	}
	resp, err := h.fileService.Upload(c.Request.Context(), reader, newName, folderID, userID, file.TeamID)
	if err != nil {
		util.InternalError(c, "copy failed")
		return
	}

	util.Created(c, resp)
}

func (h *FileHandler) Share(c *gin.Context) {
	var req dto.ShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	err := h.fileService.Share(c.Request.Context(), c.Param("id"), userID, &req)
	if err != nil {
		util.InternalError(c, "share failed")
		return
	}

	util.Success(c, nil)
}
