package repository

import (
	"context"

	"file_sys/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type VersionRepo struct {
	db *pgxpool.Pool
}

func NewVersionRepo(db *pgxpool.Pool) *VersionRepo {
	return &VersionRepo{db: db}
}

func (r *VersionRepo) Create(ctx context.Context, v *model.FileVersion) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO file_versions (file_id, version_number, storage_key, file_size, content_hash, created_by, change_note)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at`,
		v.FileID, v.VersionNumber, v.StorageKey, v.FileSize, v.ContentHash, v.CreatedBy, v.ChangeNote,
	).Scan(&v.ID, &v.CreatedAt)
}

func (r *VersionRepo) FindByFileID(ctx context.Context, fileID string) ([]model.FileVersion, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, file_id, version_number, storage_key, file_size, content_hash, created_by, change_note, created_at
		 FROM file_versions WHERE file_id = $1
		 ORDER BY version_number DESC`, fileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []model.FileVersion
	for rows.Next() {
		var v model.FileVersion
		if err := rows.Scan(&v.ID, &v.FileID, &v.VersionNumber, &v.StorageKey,
			&v.FileSize, &v.ContentHash, &v.CreatedBy, &v.ChangeNote, &v.CreatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, nil
}

func (r *VersionRepo) FindByID(ctx context.Context, id string) (*model.FileVersion, error) {
	v := &model.FileVersion{}
	err := r.db.QueryRow(ctx,
		`SELECT id, file_id, version_number, storage_key, file_size, content_hash, created_by, change_note, created_at
		 FROM file_versions WHERE id = $1`, id,
	).Scan(&v.ID, &v.FileID, &v.VersionNumber, &v.StorageKey,
		&v.FileSize, &v.ContentHash, &v.CreatedBy, &v.ChangeNote, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (r *VersionRepo) GetNextVersionNumber(ctx context.Context, fileID string) (int, error) {
	var maxVersion int
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(MAX(version_number), 0) FROM file_versions WHERE file_id = $1`, fileID,
	).Scan(&maxVersion)
	if err != nil {
		return 0, err
	}
	return maxVersion + 1, nil
}
