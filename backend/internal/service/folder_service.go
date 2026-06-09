package service

import (
	"context"
	"strings"

	"file_sys/backend/internal/dto"
	"file_sys/backend/internal/model"
	"file_sys/backend/internal/repository"
)

type FolderService struct {
	folderRepo *repository.FolderRepo
	permRepo   *repository.PermissionRepo
}

func NewFolderService(folderRepo *repository.FolderRepo, permRepo *repository.PermissionRepo) *FolderService {
	return &FolderService{folderRepo: folderRepo, permRepo: permRepo}
}

func (s *FolderService) Create(ctx context.Context, req *dto.CreateFolderRequest, ownerID string) (*model.Folder, error) {
	folder := &model.Folder{
		Name:     req.Name,
		ParentID: req.ParentID,
		OwnerID:  ownerID,
		TeamID:   req.TeamID,
	}
	if err := s.folderRepo.Create(ctx, folder); err != nil {
		return nil, err
	}

	// Compute ltree folder_path: parent_path + "." + own_id, or just own_id for root
	folderPath := folder.ID
	if req.ParentID != nil && *req.ParentID != "" {
		parent, err := s.folderRepo.FindByID(ctx, *req.ParentID)
		if err == nil && parent.FolderPath != "" {
			folderPath = parent.FolderPath + "." + folder.ID
		}
	}
	_ = s.folderRepo.SetPath(ctx, folder.ID, folderPath)
	folder.FolderPath = folderPath

	return folder, nil
}

func (s *FolderService) GetByID(ctx context.Context, id string) (*dto.FolderDetailResponse, error) {
	f, err := s.folderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := &dto.FolderDetailResponse{
		ID:         f.ID,
		Name:       f.Name,
		ParentID:   f.ParentID,
		OwnerID:    f.OwnerID,
		TeamID:     f.TeamID,
		FolderPath: f.FolderPath,
		IsDeleted:  f.IsDeleted,
		DeletedAt:  f.DeletedAt,
		CreatedAt:  f.CreatedAt,
		UpdatedAt:  f.UpdatedAt,
	}

	// Build breadcrumb from ancestor IDs in folder_path
	if f.FolderPath != "" {
		parts := strings.Split(f.FolderPath, ".")
		ancestorIDs := make([]string, 0, len(parts))
		for _, p := range parts {
			if p != "" {
				ancestorIDs = append(ancestorIDs, p)
			}
		}
		if len(ancestorIDs) > 0 {
			ancestors, err := s.folderRepo.FindByIDs(ctx, ancestorIDs)
			if err == nil {
				resp.Breadcrumb = make([]dto.FolderBreadcrumbItem, 0, len(ancestors))
				for _, a := range ancestors {
					resp.Breadcrumb = append(resp.Breadcrumb, dto.FolderBreadcrumbItem{
						ID: a.ID, Name: a.Name,
					})
				}
			}
		}
	}
	// Include current folder as last breadcrumb item
	resp.Breadcrumb = append(resp.Breadcrumb, dto.FolderBreadcrumbItem{
		ID: f.ID, Name: f.Name,
	})

	return resp, nil
}

func (s *FolderService) ListByParent(ctx context.Context, parentID *string, ownerID string, page, pageSize int) ([]model.Folder, int64, error) {
	return s.folderRepo.FindByParent(ctx, parentID, ownerID, page, pageSize)
}

func (s *FolderService) ListByTeam(ctx context.Context, teamID string, parentID *string, page, pageSize int) ([]model.Folder, int64, error) {
	return s.folderRepo.FindByTeam(ctx, teamID, parentID, page, pageSize)
}

func (s *FolderService) GetAllByOwner(ctx context.Context, ownerID string) ([]model.Folder, error) {
	return s.folderRepo.FindByOwner(ctx, ownerID)
}

func (s *FolderService) Update(ctx context.Context, id string, req *dto.UpdateFolderRequest) error {
	return s.folderRepo.Update(ctx, id, req.Name, req.ParentID)
}

func (s *FolderService) Delete(ctx context.Context, id string) error {
	return s.folderRepo.SoftDelete(ctx, id)
}

func (s *FolderService) Share(ctx context.Context, folderID, grantedBy string, req *dto.ShareRequest) error {
	perm := &model.Permission{
		FolderID:    &folderID,
		UserID:      req.UserID,
		TeamID:      req.TeamID,
		Permission:  req.Permission,
		GrantedBy:  grantedBy,
	}
	return s.permRepo.Create(ctx, perm)
}

func (s *FolderService) FindDeleted(ctx context.Context, ownerID string) ([]model.Folder, error) {
	return s.folderRepo.FindDeleted(ctx, ownerID)
}

func (s *FolderService) FindDeletedByTeam(ctx context.Context, teamID string) ([]model.Folder, error) {
	return s.folderRepo.FindDeletedByTeam(ctx, teamID)
}

func (s *FolderService) Restore(ctx context.Context, id string) error {
	return s.folderRepo.Restore(ctx, id)
}

func (s *FolderService) PermanentDelete(ctx context.Context, id string) error {
	return s.folderRepo.PermanentDelete(ctx, id)
}
