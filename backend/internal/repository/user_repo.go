package repository

import (
	"context"
	"time"

	"file_sys/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, display_name, role)
		 VALUES ($1, $2, $3, COALESCE($4, 'member'))
		 RETURNING id, created_at, updated_at`,
		user.Email, user.PasswordHash, user.DisplayName, user.Role,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, display_name, avatar_url, role, is_active, last_login_at, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName,
		&user.AvatarURL, &user.Role, &user.IsActive, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, display_name, avatar_url, role, is_active, last_login_at, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName,
		&user.AvatarURL, &user.Role, &user.IsActive, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) FindByIDs(ctx context.Context, ids []string) (map[string]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, display_name FROM users WHERE id = ANY($1)`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]string, len(ids))
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		m[id] = name
	}
	return m, nil
}

func (r *UserRepo) UpdateLastLogin(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET last_login_at = $1, updated_at = $1 WHERE id = $2`,
		time.Now(), userID)
	return err
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`,
		passwordHash, time.Now(), userID)
	return err
}

func (r *UserRepo) FindAll(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
	var total int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE is_active = true`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.Query(ctx,
		`SELECT id, email, display_name, avatar_url, role, is_active, last_login_at, created_at, updated_at
		 FROM users WHERE is_active = true
		 ORDER BY created_at DESC LIMIT $1 OFFSET $2`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL,
			&u.Role, &u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}
