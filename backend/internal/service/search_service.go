package service

import (
	"context"

	"file_sys/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SearchService struct {
	db *pgxpool.Pool
}

func NewSearchService(db *pgxpool.Pool) *SearchService {
	return &SearchService{db: db}
}

func (s *SearchService) Search(ctx context.Context, userID, query string) ([]model.File, []model.Folder, error) {
	// Search files by name using trigram similarity
	fileRows, err := s.db.Query(ctx,
		`SELECT id, name, original_name, folder_id, owner_id, mime_type, file_size,
		        file_ext, content_hash, current_version, is_deleted, created_at, updated_at
		 FROM files
		 WHERE owner_id = $1 AND is_deleted = false
		   AND (name ILIKE '%' || $2 || '%' OR original_name ILIKE '%' || $2 || '%')
		 ORDER BY updated_at DESC LIMIT 50`, userID, query)
	if err != nil {
		return nil, nil, err
	}
	defer fileRows.Close()

	var files []model.File
	for fileRows.Next() {
		var f model.File
		if err := fileRows.Scan(&f.ID, &f.Name, &f.OriginalName, &f.FolderID, &f.OwnerID,
			&f.MimeType, &f.FileSize, &f.FileExt, &f.ContentHash,
			&f.CurrentVersion, &f.IsDeleted, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, nil, err
		}
		files = append(files, f)
	}

	// Search folders
	folderRows, err := s.db.Query(ctx,
		`SELECT id, name, parent_id, owner_id, team_id, folder_path, is_deleted, created_at, updated_at
		 FROM folders
		 WHERE owner_id = $1 AND is_deleted = false
		   AND name ILIKE '%' || $2 || '%'
		 ORDER BY updated_at DESC LIMIT 20`, userID, query)
	if err != nil {
		return files, nil, err
	}
	defer folderRows.Close()

	var folders []model.Folder
	for folderRows.Next() {
		var f model.Folder
		if err := folderRows.Scan(&f.ID, &f.Name, &f.ParentID, &f.OwnerID, &f.TeamID,
			&f.FolderPath, &f.IsDeleted, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return files, folders, err
		}
		folders = append(folders, f)
	}

	return files, folders, nil
}
