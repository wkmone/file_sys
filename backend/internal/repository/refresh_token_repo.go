package repository

import (
	"context"
	"time"

	"file_sys/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokenRepo struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepo(db *pgxpool.Pool) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, token *model.RefreshToken) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, device_info, ip_address, expires_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		token.UserID, token.TokenHash, token.DeviceInfo, token.IPAddress, token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)
}

func (r *RefreshTokenRepo) FindByHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	token := &model.RefreshToken{}
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, token_hash, device_info, ip_address, expires_at, revoked, created_at
		 FROM refresh_tokens WHERE token_hash = $1`, hash,
	).Scan(&token.ID, &token.UserID, &token.TokenHash, &token.DeviceInfo,
		&token.IPAddress, &token.ExpiresAt, &token.Revoked, &token.CreatedAt)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *RefreshTokenRepo) Revoke(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = true WHERE id = $1`, id)
	return err
}

func (r *RefreshTokenRepo) RevokeAllUserTokens(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = true WHERE user_id = $1 AND revoked = false`, userID)
	return err
}

func (r *RefreshTokenRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE expires_at < $1`, time.Now())
	return err
}
