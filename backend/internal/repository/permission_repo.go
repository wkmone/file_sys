package repository

import (
	"context"

	"file_sys/backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PermissionRepo struct {
	db *pgxpool.Pool
}

func NewPermissionRepo(db *pgxpool.Pool) *PermissionRepo {
	return &PermissionRepo{db: db}
}

func (r *PermissionRepo) Create(ctx context.Context, p *model.Permission) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO permissions (folder_id, file_id, user_id, team_id, permission, granted_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		p.FolderID, p.FileID, p.UserID, p.TeamID, p.Permission, p.GrantedBy,
	).Scan(&p.ID, &p.CreatedAt)
}

func (r *PermissionRepo) FindByFolder(ctx context.Context, folderID string) ([]model.Permission, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, folder_id, file_id, user_id, team_id, permission, granted_by, created_at
		 FROM permissions WHERE folder_id = $1`, folderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []model.Permission
	for rows.Next() {
		var p model.Permission
		if err := rows.Scan(&p.ID, &p.FolderID, &p.FileID, &p.UserID, &p.TeamID,
			&p.Permission, &p.GrantedBy, &p.CreatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

func (r *PermissionRepo) FindByFile(ctx context.Context, fileID string) ([]model.Permission, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, folder_id, file_id, user_id, team_id, permission, granted_by, created_at
		 FROM permissions WHERE file_id = $1`, fileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []model.Permission
	for rows.Next() {
		var p model.Permission
		if err := rows.Scan(&p.ID, &p.FolderID, &p.FileID, &p.UserID, &p.TeamID,
			&p.Permission, &p.GrantedBy, &p.CreatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

func (r *PermissionRepo) FindByUserAndFile(ctx context.Context, userID, fileID string) (*model.Permission, error) {
	p := &model.Permission{}
	err := r.db.QueryRow(ctx,
		`SELECT id, folder_id, file_id, user_id, team_id, permission, granted_by, created_at
		 FROM permissions WHERE user_id = $1 AND file_id = $2`, userID, fileID,
	).Scan(&p.ID, &p.FolderID, &p.FileID, &p.UserID, &p.TeamID,
		&p.Permission, &p.GrantedBy, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PermissionRepo) FindByUserAndFolder(ctx context.Context, userID, folderID string) (*model.Permission, error) {
	p := &model.Permission{}
	err := r.db.QueryRow(ctx,
		`SELECT id, folder_id, file_id, user_id, team_id, permission, granted_by, created_at
		 FROM permissions WHERE user_id = $1 AND folder_id = $2`, userID, folderID,
	).Scan(&p.ID, &p.FolderID, &p.FileID, &p.UserID, &p.TeamID,
		&p.Permission, &p.GrantedBy, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

type PermissionInfo struct {
	ID         string  `json:"id"`
	UserID     *string `json:"user_id"`
	UserName   string  `json:"user_name"`
	TeamID     *string `json:"team_id"`
	TeamName   string  `json:"team_name"`
	Permission string  `json:"permission"`
	GrantedBy  string  `json:"granted_by"`
	CreatedAt  string  `json:"created_at"`
}

func (r *PermissionRepo) FindByResource(ctx context.Context, resourceType, resourceID string) ([]PermissionInfo, error) {
	var rows pgx.Rows
	var err error
	if resourceType == "file" {
		rows, err = r.db.Query(ctx,
			`SELECT p.id, p.user_id, COALESCE(u.display_name, ''), p.team_id, COALESCE(t.name, ''),
			        p.permission, p.granted_by, p.created_at
			 FROM permissions p
			 LEFT JOIN users u ON p.user_id = u.id
			 LEFT JOIN teams t ON p.team_id = t.id
			 WHERE p.file_id = $1`, resourceID)
	} else {
		rows, err = r.db.Query(ctx,
			`SELECT p.id, p.user_id, COALESCE(u.display_name, ''), p.team_id, COALESCE(t.name, ''),
			        p.permission, p.granted_by, p.created_at
			 FROM permissions p
			 LEFT JOIN users u ON p.user_id = u.id
			 LEFT JOIN teams t ON p.team_id = t.id
			 WHERE p.folder_id = $1`, resourceID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []PermissionInfo
	for rows.Next() {
		var pi PermissionInfo
		if err := rows.Scan(&pi.ID, &pi.UserID, &pi.UserName, &pi.TeamID, &pi.TeamName,
			&pi.Permission, &pi.GrantedBy, &pi.CreatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, pi)
	}
	return perms, nil
}

func (r *PermissionRepo) FindSharedFilesByUser(ctx context.Context, userID string) ([]model.File, error) {
	rows, err := r.db.Query(ctx,
		`SELECT DISTINCT f.id, f.name, f.original_name, f.folder_id, f.owner_id, f.team_id,
		        f.mime_type, f.file_size, f.file_ext, f.storage_key, f.content_hash,
		        f.current_version, f.is_deleted, f.deleted_at, f.created_at, f.updated_at
		 FROM files f
		 INNER JOIN permissions p ON f.id = p.file_id AND p.user_id = $1
		 WHERE f.is_deleted = false
		 ORDER BY f.updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []model.File
	for rows.Next() {
		var f model.File
		if err := rows.Scan(&f.ID, &f.Name, &f.OriginalName, &f.FolderID, &f.OwnerID, &f.TeamID,
			&f.MimeType, &f.FileSize, &f.FileExt, &f.StorageKey,
			&f.ContentHash, &f.CurrentVersion, &f.IsDeleted, &f.DeletedAt,
			&f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

func (r *PermissionRepo) FindSharedFoldersByUser(ctx context.Context, userID string) ([]model.Folder, error) {
	rows, err := r.db.Query(ctx,
		`SELECT DISTINCT fo.id, fo.name, fo.parent_id, fo.owner_id, fo.team_id,
		        fo.folder_path, fo.is_deleted, fo.deleted_at, fo.created_at, fo.updated_at
		 FROM folders fo
		 INNER JOIN permissions p ON fo.id = p.folder_id AND p.user_id = $1
		 WHERE fo.is_deleted = false
		 ORDER BY fo.updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []model.Folder
	for rows.Next() {
		var f model.Folder
		if err := rows.Scan(&f.ID, &f.Name, &f.ParentID, &f.OwnerID, &f.TeamID,
			&f.FolderPath, &f.IsDeleted, &f.DeletedAt, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		folders = append(folders, f)
	}
	return folders, nil
}

func (r *PermissionRepo) Update(ctx context.Context, id, permission string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE permissions SET permission = $2 WHERE id = $1`, id, permission)
	return err
}

func (r *PermissionRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM permissions WHERE id = $1`, id)
	return err
}
