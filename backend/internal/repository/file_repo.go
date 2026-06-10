package repository

import (
	"context"
	"time"

	"file_sys/backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FileRepo struct {
	db *pgxpool.Pool
}

func NewFileRepo(db *pgxpool.Pool) *FileRepo {
	return &FileRepo{db: db}
}

// Common column list used by queries that return File rows (unqualified).
const fileCols = `id, name, original_name, folder_id, owner_id, team_id, mime_type, file_size,
                   file_ext, storage_key, content_hash, current_version, is_deleted, deleted_at, created_at, updated_at`
 
// fileColsF is the same list qualified with f. for JOIN queries.
const fileColsF = `f.id, f.name, f.original_name, f.folder_id, f.owner_id, f.team_id, f.mime_type, f.file_size,
	                   f.file_ext, f.storage_key, f.content_hash, f.current_version, f.is_deleted, f.deleted_at, f.created_at, f.updated_at`

func scanFile(row pgx.Row) (*model.File, error) {
	var f model.File
	err := row.Scan(&f.ID, &f.Name, &f.OriginalName, &f.FolderID, &f.OwnerID, &f.TeamID,
		&f.MimeType, &f.FileSize, &f.FileExt, &f.StorageKey,
		&f.ContentHash, &f.CurrentVersion, &f.IsDeleted, &f.DeletedAt,
		&f.CreatedAt, &f.UpdatedAt)
	return &f, err
}

func scanFiles(rows pgx.Rows) ([]model.File, error) {
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

func scanFilesWithPerm(rows pgx.Rows) ([]model.File, error) {
	defer rows.Close()
	var files []model.File
	for rows.Next() {
		var f model.File
		if err := rows.Scan(&f.ID, &f.Name, &f.OriginalName, &f.FolderID, &f.OwnerID, &f.TeamID,
			&f.MimeType, &f.FileSize, &f.FileExt, &f.StorageKey,
			&f.ContentHash, &f.CurrentVersion, &f.IsDeleted, &f.DeletedAt,
			&f.CreatedAt, &f.UpdatedAt, &f.Permission, &f.SharedBy); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

func (r *FileRepo) Create(ctx context.Context, file *model.File) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO files (name, original_name, folder_id, owner_id, team_id, mime_type, file_size, file_ext, storage_key, content_hash, current_version)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id, created_at, updated_at`,
		file.Name, file.OriginalName, file.FolderID, file.OwnerID, file.TeamID,
		file.MimeType, file.FileSize, file.FileExt, file.StorageKey,
		file.ContentHash, file.CurrentVersion,
	).Scan(&file.ID, &file.CreatedAt, &file.UpdatedAt)
}

func (r *FileRepo) FindByID(ctx context.Context, id string) (*model.File, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+fileCols+` FROM files WHERE id = $1`, id)
	return scanFile(row)
}

func (r *FileRepo) FindByFolder(ctx context.Context, folderID *string, ownerID string, page, pageSize int) ([]model.File, int64, error) {
	var total int64
	var err error

	if folderID == nil || *folderID == "" {
		err = r.db.QueryRow(ctx,
			`SELECT COUNT(DISTINCT f.id) FROM files f
			 LEFT JOIN permissions p ON f.id = p.file_id AND p.user_id = $1::uuid
			 WHERE f.folder_id IS NULL AND f.team_id IS NULL AND f.is_deleted = false
			   AND (f.owner_id = $1 OR p.id IS NOT NULL)`, ownerID,
		).Scan(&total)
	} else {
		err = r.db.QueryRow(ctx,
			`SELECT COUNT(DISTINCT f.id) FROM files f
			 LEFT JOIN permissions p ON f.id = p.file_id AND p.user_id = $2::uuid
			 WHERE f.folder_id = $1 AND f.team_id IS NULL AND f.is_deleted = false
			   AND (f.owner_id = $2 OR p.id IS NOT NULL)`, *folderID, ownerID,
		).Scan(&total)
	}
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var rows pgx.Rows
	if folderID == nil || *folderID == "" {
		rows, err = r.db.Query(ctx,
			`SELECT DISTINCT `+fileColsF+`, COALESCE(p.permission, '') AS perm, COALESCE(p.granted_by::text, '') AS shared_by
			 FROM files f
			 LEFT JOIN permissions p ON f.id = p.file_id AND p.user_id = $1::uuid
			 WHERE f.folder_id IS NULL AND f.team_id IS NULL AND f.is_deleted = false
			   AND (f.owner_id = $1 OR p.id IS NOT NULL)
			 ORDER BY f.name ASC LIMIT $2 OFFSET $3`, ownerID, pageSize, offset)
	} else {
		rows, err = r.db.Query(ctx,
			`SELECT DISTINCT `+fileColsF+`, COALESCE(p.permission, '') AS perm, COALESCE(p.granted_by::text, '') AS shared_by
			 FROM files f
			 LEFT JOIN permissions p ON f.id = p.file_id AND p.user_id = $2::uuid
			 WHERE f.folder_id = $1 AND f.team_id IS NULL AND f.is_deleted = false
			   AND (f.owner_id = $2 OR p.id IS NOT NULL)
			 ORDER BY f.name ASC LIMIT $3 OFFSET $4`, *folderID, ownerID, pageSize, offset)
	}
	if err != nil {
		return nil, 0, err
	}

	files, err := scanFilesWithPerm(rows)
	return files, total, err
}

func (r *FileRepo) Update(ctx context.Context, id string, name *string, folderID *string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE files SET
			name = COALESCE($2, name),
			folder_id = COALESCE($3, folder_id),
			updated_at = $4
		 WHERE id = $1`,
		id, name, folderID, time.Now())
	return err
}

func (r *FileRepo) UpdateVersion(ctx context.Context, id string, storageKey string, contentHash string, newVersion int, fileSize int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE files SET storage_key = $2, content_hash = $3, current_version = $4, file_size = $5, updated_at = $6
		 WHERE id = $1`,
		id, storageKey, contentHash, newVersion, fileSize, time.Now())
	return err
}

func (r *FileRepo) SoftDelete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE files SET is_deleted = true, deleted_at = $2, updated_at = $2 WHERE id = $1`,
		id, time.Now())
	return err
}

func (r *FileRepo) FindDeleted(ctx context.Context, ownerID string) ([]model.File, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+fileCols+` FROM files WHERE owner_id = $1 AND team_id IS NULL AND is_deleted = true
		 ORDER BY deleted_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	return scanFiles(rows)
}

func (r *FileRepo) FindDeletedByTeam(ctx context.Context, teamID string) ([]model.File, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+fileCols+` FROM files WHERE team_id = $1 AND is_deleted = true
		 ORDER BY deleted_at DESC`, teamID)
	if err != nil {
		return nil, err
	}
	return scanFiles(rows)
}

func (r *FileRepo) Restore(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE files SET is_deleted = false, deleted_at = NULL, updated_at = $2 WHERE id = $1`,
		id, time.Now())
	return err
}

func (r *FileRepo) PermanentDelete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM files WHERE id = $1`, id)
	return err
}

// FindByTeam finds files that belong to the given team, either directly (root level)
// or via a team-owned folder.
func (r *FileRepo) FindByTeam(ctx context.Context, teamID string, folderID *string, userID string, page, pageSize int) ([]model.File, int64, error) {
	var total int64
	var rows pgx.Rows
	var err error

	if folderID == nil || *folderID == "" {
		err = r.db.QueryRow(ctx,
			`SELECT COUNT(*) FROM files f
			 WHERE f.folder_id IS NULL AND f.is_deleted = false
			   AND (f.team_id = $1 OR EXISTS (
			       SELECT 1 FROM permissions p WHERE p.file_id = f.id AND p.user_id = $2
			   ))`, teamID, userID,
		).Scan(&total)
	} else {
		err = r.db.QueryRow(ctx,
			`SELECT COUNT(*) FROM files f
			 WHERE f.folder_id = $1 AND f.is_deleted = false
			   AND (f.team_id = $2 OR EXISTS (
			       SELECT 1 FROM permissions p WHERE p.file_id = f.id AND p.user_id = $3
			   ))`, *folderID, teamID, userID,
		).Scan(&total)
	}
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if folderID == nil || *folderID == "" {
		rows, err = r.db.Query(ctx,
			`SELECT `+fileCols+` FROM files f
			 WHERE f.folder_id IS NULL AND f.is_deleted = false
			   AND (f.team_id = $1 OR EXISTS (
			       SELECT 1 FROM permissions p WHERE p.file_id = f.id AND p.user_id = $2
			   ))
			 ORDER BY f.updated_at DESC LIMIT $3 OFFSET $4`, teamID, userID, pageSize, offset)
	} else {
		rows, err = r.db.Query(ctx,
			`SELECT `+fileCols+` FROM files f
			 WHERE f.folder_id = $1 AND f.is_deleted = false
			   AND (f.team_id = $2 OR EXISTS (
			       SELECT 1 FROM permissions p WHERE p.file_id = f.id AND p.user_id = $3
			   ))
			 ORDER BY f.updated_at DESC LIMIT $4 OFFSET $5`, *folderID, teamID, userID, pageSize, offset)
	}
	if err != nil {
		return nil, 0, err
	}

	files, err := scanFiles(rows)
	return files, total, err
}
