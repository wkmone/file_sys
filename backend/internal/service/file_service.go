package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"file_sys/backend/internal/dto"
	"file_sys/backend/internal/model"
	"file_sys/backend/internal/repository"
	"file_sys/backend/internal/storage"
)

type FileService struct {
	fileRepo      *repository.FileRepo
	versionRepo   *repository.VersionRepo
	folderRepo    *repository.FolderRepo
	permRepo      *repository.PermissionRepo
	userRepo      *repository.UserRepo
	store         storage.Storage
}

func NewFileService(
	fileRepo *repository.FileRepo,
	versionRepo *repository.VersionRepo,
	folderRepo *repository.FolderRepo,
	permRepo *repository.PermissionRepo,
	userRepo *repository.UserRepo,
	store storage.Storage,
) *FileService {
	return &FileService{
		fileRepo:    fileRepo,
		versionRepo: versionRepo,
		folderRepo:  folderRepo,
		permRepo:    permRepo,
		userRepo:    userRepo,
		store:       store,
	}
}

func (s *FileService) Upload(ctx context.Context, reader io.Reader, originalName, folderID, ownerID string, teamID *string) (*dto.UploadFileResponse, error) {
	ext := strings.ToLower(filepath.Ext(originalName))
	mimeType := detectMimeType(ext)

	// Write to temp file while computing hash (single pass)
	tmpFile, err := os.CreateTemp("", "upload_*.tmp")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	hash, err := storage.HashStream(reader, tmpFile)
	if err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("hash stream: %w", err)
	}
	fileSize, _ := tmpFile.Seek(0, io.SeekEnd)
	tmpFile.Close()

	key := storage.StorageKey(hash)

	// Dedup: if same hash already stored, skip upload
	exists, _ := s.store.Exists(ctx, key)
	if !exists {
		tmpFile, err = os.Open(tmpFile.Name())
		if err != nil {
			return nil, fmt.Errorf("reopen temp file: %w", err)
		}
		defer tmpFile.Close()
		if err := s.store.Store(ctx, key, tmpFile, mimeType); err != nil {
			return nil, fmt.Errorf("store file: %w", err)
		}
	}

	file := &model.File{
		Name:           originalName,
		OriginalName:   originalName,
		FolderID:       nilIfEmpty(folderID),
		OwnerID:        ownerID,
		TeamID:         teamID,
		MimeType:       mimeType,
		FileSize:       fileSize,
		FileExt:        ext,
		StorageKey:     key,
		ContentHash:    hash,
		CurrentVersion: 1,
	}
	if err := s.fileRepo.Create(ctx, file); err != nil {
		return nil, err
	}

	version := &model.FileVersion{
		FileID:        file.ID,
		VersionNumber: 1,
		StorageKey:    key,
		FileSize:      fileSize,
		ContentHash:   hash,
		CreatedBy:     &ownerID,
	}
	if err := s.versionRepo.Create(ctx, version); err != nil {
		return nil, err
	}

	return &dto.UploadFileResponse{
		ID:             file.ID,
		Name:           file.Name,
		OriginalName:   file.OriginalName,
		FolderID:       file.FolderID,
		OwnerID:        file.OwnerID,
		MimeType:       file.MimeType,
		FileSize:       file.FileSize,
		FileExt:        file.FileExt,
		CurrentVersion: file.CurrentVersion,
		CreatedAt:      file.CreatedAt,
		UpdatedAt:      file.UpdatedAt,
	}, nil
}

func (s *FileService) BatchUpload(ctx context.Context, files []*multipart.FileHeader, manifestJSON string, folderID string, ownerID string, teamID *string) (*dto.BatchUploadResponse, error) {
	var manifest []dto.BatchUploadManifestEntry
	if err := json.Unmarshal([]byte(manifestJSON), &manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest JSON: %w", err)
	}

	// Collect unique directory paths from all entries
	dirSet := make(map[string]bool)
	for _, entry := range manifest {
		dir := path.Dir(entry.Path)
		if dir != "." {
			parts := strings.Split(dir, "/")
			for i := range parts {
				dirSet[strings.Join(parts[:i+1], "/")] = true
			}
		}
		if entry.IsDirectory {
			dirSet[entry.Path] = true
		}
	}

	// Sort by depth so parents are created before children
	dirs := make([]string, 0, len(dirSet))
	for d := range dirSet {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)

	// Create folder hierarchy: path -> folderID
	folderCache := make(map[string]string)
	for _, dir := range dirs {
		parentPath := path.Dir(dir)
		var parentID *string
		if parentPath != "." {
			pid := folderCache[parentPath]
			parentID = &pid
		} else if folderID != "" {
			parentID = &folderID
		}

		// Try to find existing folder
		existing, err := s.folderRepo.FindByNameAndParent(ctx, filepath.Base(dir), parentID, ownerID, teamID)
		if err == nil {
			folderCache[dir] = existing.ID
			continue
		}

		// Create new folder
		folder := &model.Folder{
			Name:     filepath.Base(dir),
			ParentID: parentID,
			OwnerID:  ownerID,
			TeamID:   teamID,
		}
		if err := s.folderRepo.Create(ctx, folder); err != nil {
			return nil, fmt.Errorf("create folder %s: %w", dir, err)
		}
			// Compute ltree folder_path
			fp := folder.ID
			if parentID != nil && *parentID != "" {
				parent, perr := s.folderRepo.FindByID(ctx, *parentID)
				if perr == nil && parent.FolderPath != "" {
					fp = parent.FolderPath + "." + folder.ID
				}
			}
			_ = s.folderRepo.SetPath(ctx, folder.ID, fp)
			folderCache[dir] = folder.ID
		}

	// Upload files, matching by position (non-directory entries only)
	results := make([]dto.BatchUploadFileResult, 0, len(manifest))
	succeeded, failed := 0, 0
	fileIdx := 0

	for _, entry := range manifest {
		if entry.IsDirectory {
			results = append(results, dto.BatchUploadFileResult{
				Path:   entry.Path,
				Status: "success",
			})
			succeeded++
			continue
		}

		result := dto.BatchUploadFileResult{Path: entry.Path}

		// Determine target folder
		dir := path.Dir(entry.Path)
		var targetFolderID string
		if dir != "." {
			targetFolderID = folderCache[dir]
		} else {
			targetFolderID = folderID
		}

		if fileIdx >= len(files) {
			result.Status = "failed"
			result.Error = "file missing in request"
			results = append(results, result)
			failed++
			continue
		}

		fh := files[fileIdx]
		fileIdx++

		reader, err := fh.Open()
		if err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			results = append(results, result)
			failed++
			continue
		}

		uploadResp, err := s.Upload(ctx, reader, filepath.Base(entry.Path), targetFolderID, ownerID, teamID)
		reader.(io.Closer).Close()
		if err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			failed++
		} else {
			result.Status = "success"
			result.FileID = uploadResp.ID
			succeeded++
		}
		results = append(results, result)
	}

	return &dto.BatchUploadResponse{
		Total:     len(manifest),
		Succeeded: succeeded,
		Failed:    failed,
		Results:   results,
	}, nil
}

func (s *FileService) GetByID(ctx context.Context, id string) (*model.File, error) {
	return s.fileRepo.FindByID(ctx, id)
}

func (s *FileService) enrichOwnerNames(ctx context.Context, files []model.File) {
	ownerIDs := make([]string, 0, len(files))
	seen := make(map[string]bool, len(files))
	for _, f := range files {
		if f.OwnerID != "" && !seen[f.OwnerID] {
			ownerIDs = append(ownerIDs, f.OwnerID)
			seen[f.OwnerID] = true
		}
	}
	if len(ownerIDs) == 0 {
		return
	}
	names, err := s.userRepo.FindByIDs(ctx, ownerIDs)
	if err != nil {
		return
	}
	for i := range files {
		if n, ok := names[files[i].OwnerID]; ok {
			files[i].OwnerName = n
		}
	}
}

func (s *FileService) ListByFolder(ctx context.Context, folderID *string, ownerID string, page, pageSize int) ([]model.File, int64, error) {
	files, total, err := s.fileRepo.FindByFolder(ctx, folderID, ownerID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	s.enrichOwnerNames(ctx, files)
	return files, total, nil
}

func (s *FileService) ListByTeam(ctx context.Context, teamID string, folderID *string, page, pageSize int) ([]model.File, int64, error) {
	files, total, err := s.fileRepo.FindByTeam(ctx, teamID, folderID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	s.enrichOwnerNames(ctx, files)
	return files, total, nil
}

func (s *FileService) Update(ctx context.Context, id string, req *dto.UpdateFileRequest) error {
	return s.fileRepo.Update(ctx, id, req.Name, req.FolderID)
}

func (s *FileService) Delete(ctx context.Context, id string) error {
	return s.fileRepo.SoftDelete(ctx, id)
}

func (s *FileService) GetStorageKey(ctx context.Context, id string) (string, error) {
	file, err := s.fileRepo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	return file.StorageKey, nil
}

func (s *FileService) GetVersionStorageKey(ctx context.Context, versionID string) (string, error) {
	v, err := s.versionRepo.FindByID(ctx, versionID)
	if err != nil {
		return "", err
	}
	return v.StorageKey, nil
}

func (s *FileService) CreateNewVersion(ctx context.Context, fileID string, reader io.Reader, createdBy string, note string) (*model.FileVersion, error) {
	var buf bytes.Buffer
	hash, err := storage.HashStream(reader, &buf)
	if err != nil {
		return nil, err
	}

	key := storage.StorageKey(hash)

	exists, _ := s.store.Exists(ctx, key)
	if !exists {
		if err := s.store.Store(ctx, key, &buf, "application/octet-stream"); err != nil {
			return nil, fmt.Errorf("store new version: %w", err)
		}
	}

	nextVersion, err := s.versionRepo.GetNextVersionNumber(ctx, fileID)
	if err != nil {
		return nil, err
	}

	version := &model.FileVersion{
		FileID:        fileID,
		VersionNumber: nextVersion,
		StorageKey:    key,
		ContentHash:   hash,
		CreatedBy:     &createdBy,
		ChangeNote:    &note,
	}
	if err := s.versionRepo.Create(ctx, version); err != nil {
		return nil, err
	}

	// Update file record
	_ = s.fileRepo.UpdateVersion(ctx, fileID, key, hash, nextVersion, 0)

	return version, nil
}

func (s *FileService) Share(ctx context.Context, fileID, grantedBy string, req *dto.ShareRequest) error {
	perm := &model.Permission{
		FileID:     &fileID,
		UserID:      req.UserID,
		TeamID:      req.TeamID,
		Permission:  req.Permission,
		GrantedBy:  grantedBy,
	}
	return s.permRepo.Create(ctx, perm)
}

func (s *FileService) FindDeleted(ctx context.Context, ownerID string) ([]model.File, error) {
	return s.fileRepo.FindDeleted(ctx, ownerID)
}

func (s *FileService) FindDeletedByTeam(ctx context.Context, teamID string) ([]model.File, error) {
	return s.fileRepo.FindDeletedByTeam(ctx, teamID)
}

func (s *FileService) Restore(ctx context.Context, id string) error {
	return s.fileRepo.Restore(ctx, id)
}

func (s *FileService) PermanentDelete(ctx context.Context, id string) error {
	file, err := s.fileRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	// Delete from storage
	s.store.Delete(ctx, file.StorageKey)
	return s.fileRepo.PermanentDelete(ctx, id)
}

func (s *FileService) GetVersions(ctx context.Context, fileID string) ([]model.FileVersion, error) {
	return s.versionRepo.FindByFileID(ctx, fileID)
}

func (s *FileService) GetVersion(ctx context.Context, versionID string) (*model.FileVersion, error) {
	return s.versionRepo.FindByID(ctx, versionID)
}

func (s *FileService) DownloadVersion(ctx context.Context, versionID string) (io.ReadCloser, *VersionDownloadInfo, error) {
	v, err := s.versionRepo.FindByID(ctx, versionID)
	if err != nil {
		return nil, nil, err
	}
	reader, fi, err := s.store.Retrieve(ctx, v.StorageKey)
	if err != nil {
		return nil, nil, err
	}
	return reader, &VersionDownloadInfo{Size: fi.Size, ContentType: "application/octet-stream"}, nil
}

type VersionDownloadInfo struct {
	Size        int64
	ContentType string
}

func (s *FileService) RestoreVersion(ctx context.Context, fileID, versionID string) error {
	v, err := s.versionRepo.FindByID(ctx, versionID)
	if err != nil {
		return err
	}
	nextVersion, _ := s.versionRepo.GetNextVersionNumber(ctx, fileID)

	// Create a new version that copies the old content
	restoredVersion := &model.FileVersion{
		FileID:        fileID,
		VersionNumber: nextVersion,
		StorageKey:    v.StorageKey,
		FileSize:      v.FileSize,
		ContentHash:   v.ContentHash,
		CreatedBy:     v.CreatedBy,
		ChangeNote:    nil,
	}
	note := "Restored from version " + string(rune(v.VersionNumber+'0'))
	restoredVersion.ChangeNote = &note

	if err := s.versionRepo.Create(ctx, restoredVersion); err != nil {
		return err
	}
	return s.fileRepo.UpdateVersion(ctx, fileID, v.StorageKey, v.ContentHash, nextVersion, v.FileSize)
}

func (s *FileService) RetrieveStorage(ctx context.Context, key string) (io.ReadCloser, *storage.FileInfo, error) {
	return s.store.Retrieve(ctx, key)
}

func detectMimeType(ext string) string {
	mimeTypes := map[string]string{
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".pdf":  "application/pdf",
		".txt":  "text/plain",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
	}
	if m, ok := mimeTypes[ext]; ok {
		return m
	}
	return "application/octet-stream"
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
