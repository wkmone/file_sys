package repository

import (
	"context"
	"time"

	"file_sys/backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FolderRepo struct {
	db *pgxpool.Pool
}

func NewFolderRepo(db *pgxpool.Pool) *FolderRepo {
	return &FolderRepo{db: db}
}

func (r *FolderRepo) Create(ctx context.Context, folder *model.Folder) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO folders (name, parent_id, owner_id, team_id, folder_path)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		folder.Name, folder.ParentID, folder.OwnerID, folder.TeamID, folder.FolderPath,
	).Scan(&folder.ID, &folder.CreatedAt, &folder.UpdatedAt)
}

func (r *FolderRepo) FindByNameAndParent(ctx context.Context, name string, parentID *string, ownerID string, teamID *string) (*model.Folder, error) {
	f := &model.Folder{}
	var err error
	if teamID != nil && *teamID != "" {
		err = r.db.QueryRow(ctx,
			`SELECT id, name, parent_id, owner_id, team_id, folder_path, is_deleted, deleted_at, created_at, updated_at
			 FROM folders WHERE name = $1 AND (parent_id IS NOT DISTINCT FROM $2) AND team_id = $3 AND is_deleted = false`,
			name, parentID, *teamID,
		).Scan(&f.ID, &f.Name, &f.ParentID, &f.OwnerID, &f.TeamID,
			&f.FolderPath, &f.IsDeleted, &f.DeletedAt, &f.CreatedAt, &f.UpdatedAt)
	} else {
		err = r.db.QueryRow(ctx,
			`SELECT id, name, parent_id, owner_id, team_id, folder_path, is_deleted, deleted_at, created_at, updated_at
			 FROM folders WHERE name = $1 AND (parent_id IS NOT DISTINCT FROM $2) AND owner_id = $3 AND team_id IS NULL AND is_deleted = false`,
			name, parentID, ownerID,
		).Scan(&f.ID, &f.Name, &f.ParentID, &f.OwnerID, &f.TeamID,
			&f.FolderPath, &f.IsDeleted, &f.DeletedAt, &f.CreatedAt, &f.UpdatedAt)
	}
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (r *FolderRepo) FindByID(ctx context.Context, id string) (*model.Folder, error) {
	f := &model.Folder{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, parent_id, owner_id, team_id, folder_path, is_deleted, deleted_at, created_at, updated_at
		 FROM folders WHERE id = $1`, id,
	).Scan(&f.ID, &f.Name, &f.ParentID, &f.OwnerID, &f.TeamID,
		&f.FolderPath, &f.IsDeleted, &f.DeletedAt, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (r *FolderRepo) FindByParent(ctx context.Context, parentID *string, ownerID string, page, pageSize int) ([]model.Folder, int64, error) {
	var total int64
	var rows pgx.Rows
	var err error

	if parentID == nil || *parentID == "" {
		err = r.db.QueryRow(ctx,
			`SELECT COUNT(*) FROM folders WHERE parent_id IS NULL AND owner_id = $1 AND team_id IS NULL AND is_deleted = false`, ownerID,
		).Scan(&total)
	} else {
		err = r.db.QueryRow(ctx,
			`SELECT COUNT(*) FROM folders WHERE parent_id = $1 AND owner_id = $2 AND team_id IS NULL AND is_deleted = false`, *parentID, ownerID,
		).Scan(&total)
	}
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if parentID == nil || *parentID == "" {
		rows, err = r.db.Query(ctx,
			`SELECT id, name, parent_id, owner_id, team_id, folder_path, is_deleted, created_at, updated_at
			 FROM folders WHERE parent_id IS NULL AND owner_id = $1 AND team_id IS NULL AND is_deleted = false
			 ORDER BY name ASC LIMIT $2 OFFSET $3`, ownerID, pageSize, offset)
	} else {
		rows, err = r.db.Query(ctx,
			`SELECT id, name, parent_id, owner_id, team_id, folder_path, is_deleted, created_at, updated_at
			 FROM folders WHERE parent_id = $1 AND owner_id = $2 AND team_id IS NULL AND is_deleted = false
			 ORDER BY name ASC LIMIT $3 OFFSET $4`, *parentID, ownerID, pageSize, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var folders []model.Folder
	for rows.Next() {
		var f model.Folder
		if err := rows.Scan(&f.ID, &f.Name, &f.ParentID, &f.OwnerID, &f.TeamID,
			&f.FolderPath, &f.IsDeleted, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, 0, err
		}
		folders = append(folders, f)
	}
	return folders, total, nil
}

func (r *FolderRepo) FindByIDs(ctx context.Context, ids []string) ([]model.Folder, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, parent_id, owner_id, team_id, folder_path, is_deleted, created_at, updated_at
		 FROM folders WHERE id = ANY($1) AND is_deleted = false`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	folderMap := make(map[string]model.Folder)
	for rows.Next() {
		var f model.Folder
		if err := rows.Scan(&f.ID, &f.Name, &f.ParentID, &f.OwnerID, &f.TeamID,
			&f.FolderPath, &f.IsDeleted, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		folderMap[f.ID] = f
	}
	// Preserve input order
	result := make([]model.Folder, 0, len(ids))
	for _, id := range ids {
		if f, ok := folderMap[id]; ok {
			result = append(result, f)
		}
	}
	return result, nil
}

func (r *FolderRepo) FindByOwner(ctx context.Context, ownerID string) ([]model.Folder, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, parent_id, owner_id, team_id, folder_path, is_deleted, created_at, updated_at
		 FROM folders WHERE owner_id = $1 AND team_id IS NULL AND is_deleted = false
		 ORDER BY name ASC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []model.Folder
	for rows.Next() {
		var f model.Folder
		if err := rows.Scan(&f.ID, &f.Name, &f.ParentID, &f.OwnerID, &f.TeamID,
			&f.FolderPath, &f.IsDeleted, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		folders = append(folders, f)
	}
	return folders, nil
}

func (r *FolderRepo) SetPath(ctx context.Context, id string, folderPath string) error {
	_, err := r.db.Exec(ctx, `UPDATE folders SET folder_path = $2 WHERE id = $1`, id, folderPath)
	return err
}

func (r *FolderRepo) Update(ctx context.Context, id string, name *string, parentID *string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE folders SET
			name = COALESCE($2, name),
			parent_id = $3,
			updated_at = $4
		 WHERE id = $1`,
		id, name, parentID, time.Now())
	return err
}

func (r *FolderRepo) SoftDelete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE folders SET is_deleted = true, deleted_at = $2, updated_at = $2 WHERE id = $1`,
		id, time.Now())
	return err
}

func (r *FolderRepo) FindDeleted(ctx context.Context, ownerID string) ([]model.Folder, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, parent_id, owner_id, team_id, folder_path, is_deleted, deleted_at, created_at, updated_at
		 FROM folders WHERE owner_id = $1 AND team_id IS NULL AND is_deleted = true
		 ORDER BY deleted_at DESC`, ownerID)
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

func (r *FolderRepo) FindDeletedByTeam(ctx context.Context, teamID string) ([]model.Folder, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, parent_id, owner_id, team_id, folder_path, is_deleted, deleted_at, created_at, updated_at
		 FROM folders WHERE team_id = $1 AND is_deleted = true
		 ORDER BY deleted_at DESC`, teamID)
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

func (r *FolderRepo) Restore(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE folders SET is_deleted = false, deleted_at = NULL, updated_at = $2 WHERE id = $1`,
		id, time.Now())
	return err
}

func (r *FolderRepo) PermanentDelete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM folders WHERE id = $1`, id)
	return err
}

func (r *FolderRepo) FindByTeam(ctx context.Context, teamID string, parentID *string, userID string, page, pageSize int) ([]model.Folder, int64, error) {
	var total int64
	var rows pgx.Rows
	var err error

	if parentID == nil || *parentID == "" {
		err = r.db.QueryRow(ctx,
			`SELECT COUNT(*) FROM folders fo
			 WHERE fo.parent_id IS NULL AND fo.is_deleted = false
			   AND (fo.team_id = $1 OR EXISTS (
			       SELECT 1 FROM permissions p WHERE p.folder_id = fo.id AND p.user_id = $2
			   ))`, teamID, userID,
		).Scan(&total)
	} else {
		err = r.db.QueryRow(ctx,
			`SELECT COUNT(*) FROM folders fo
			 WHERE fo.parent_id = $1 AND fo.is_deleted = false
			   AND (fo.team_id = $2 OR EXISTS (
			       SELECT 1 FROM permissions p WHERE p.folder_id = fo.id AND p.user_id = $3
			   ))`, *parentID, teamID, userID,
		).Scan(&total)
	}
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if parentID == nil || *parentID == "" {
		rows, err = r.db.Query(ctx,
			`SELECT fo.id, fo.name, fo.parent_id, fo.owner_id, fo.team_id,
			        fo.folder_path, fo.is_deleted, fo.created_at, fo.updated_at
			 FROM folders fo
			 WHERE fo.parent_id IS NULL AND fo.is_deleted = false
			   AND (fo.team_id = $1 OR EXISTS (
			       SELECT 1 FROM permissions p WHERE p.folder_id = fo.id AND p.user_id = $2
			   ))
			 ORDER BY fo.name ASC LIMIT $3 OFFSET $4`, teamID, userID, pageSize, offset)
	} else {
		rows, err = r.db.Query(ctx,
			`SELECT fo.id, fo.name, fo.parent_id, fo.owner_id, fo.team_id,
			        fo.folder_path, fo.is_deleted, fo.created_at, fo.updated_at
			 FROM folders fo
			 WHERE fo.parent_id = $1 AND fo.is_deleted = false
			   AND (fo.team_id = $2 OR EXISTS (
			       SELECT 1 FROM permissions p WHERE p.folder_id = fo.id AND p.user_id = $3
			   ))
			 ORDER BY fo.name ASC LIMIT $4 OFFSET $5`, *parentID, teamID, userID, pageSize, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var folders []model.Folder
	for rows.Next() {
		var f model.Folder
		if err := rows.Scan(&f.ID, &f.Name, &f.ParentID, &f.OwnerID, &f.TeamID,
			&f.FolderPath, &f.IsDeleted, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, 0, err
		}
		folders = append(folders, f)
	}
	return folders, total, nil
}
