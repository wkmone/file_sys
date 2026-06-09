package repository

import (
	"context"

	"file_sys/backend/internal/model"

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

func (r *PermissionRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM permissions WHERE id = $1`, id)
	return err
}
