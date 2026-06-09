package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"file_sys/backend/internal/dto"
	"file_sys/backend/internal/model"
	"file_sys/backend/internal/repository"
	"file_sys/backend/internal/util"

	"github.com/jackc/pgx/v5"
)

type AuthService struct {
	userRepo    *repository.UserRepo
	tokenRepo   *repository.RefreshTokenRepo
	accessSec   string
	refreshSec  string
}

func NewAuthService(userRepo *repository.UserRepo, tokenRepo *repository.RefreshTokenRepo, accessSec, refreshSec string) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		accessSec:  accessSec,
		refreshSec: refreshSec,
	}
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	if err := util.ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hash, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:        req.Email,
		PasswordHash: hash,
		DisplayName:  req.DisplayName,
		Role:         "member",
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.generateAuthResponse(ctx, user, "", "")
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest, ip, device string) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	if !util.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	s.userRepo.UpdateLastLogin(ctx, user.ID)

	return s.generateAuthResponse(ctx, user, ip, device)
}

func (s *AuthService) Refresh(ctx context.Context, rawToken string) (*dto.AuthResponse, error) {
	hash := hashToken(rawToken)
	stored, err := s.tokenRepo.FindByHash(ctx, hash)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}
	if stored.Revoked {
		// Token reuse detected — revoke all user tokens (family invalidation)
		s.tokenRepo.RevokeAllUserTokens(ctx, stored.UserID)
		return nil, errors.New("token revoked")
	}
	if time.Now().After(stored.ExpiresAt) {
		return nil, errors.New("token expired")
	}

	// Rotate: revoke old, issue new
	s.tokenRepo.Revoke(ctx, stored.ID)

	user, err := s.userRepo.FindByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}

	return s.generateAuthResponse(ctx, user, "", "")
}

func (s *AuthService) LogoutAll(ctx context.Context, userID string) error {
	return s.tokenRepo.RevokeAllUserTokens(ctx, userID)
}

func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	hash := hashToken(rawToken)
	stored, err := s.tokenRepo.FindByHash(ctx, hash)
	if err != nil {
		return nil // already invalid, no-op
	}
	return s.tokenRepo.Revoke(ctx, stored.ID)
}

func (s *AuthService) ChangePassword(ctx context.Context, userID, oldPass, newPass string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if !util.CheckPassword(oldPass, user.PasswordHash) {
		return errors.New("incorrect current password")
	}
	if err := util.ValidatePassword(newPass); err != nil {
		return err
	}
	hash, err := util.HashPassword(newPass)
	if err != nil {
		return err
	}
	s.tokenRepo.RevokeAllUserTokens(ctx, userID)
	return s.userRepo.UpdatePassword(ctx, userID, hash)
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userID string) (*model.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

func (s *AuthService) generateAuthResponse(ctx context.Context, user *model.User, ip, device string) (*dto.AuthResponse, error) {
	accessToken, err := util.GenerateAccessToken(user.ID, user.Email, user.Role, s.accessSec)
	if err != nil {
		return nil, err
	}

	claims, err := util.ValidateAccessToken(accessToken, s.accessSec)
	if err != nil {
		return nil, err
	}
	expiresIn := int(time.Until(claims.ExpiresAt.Time).Seconds())
	if expiresIn < 1 {
		expiresIn = 1
	}

	refreshRaw, err := util.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshHash := hashToken(refreshRaw)
	deviceInfo := &device
	ipAddr := &ip
	if device == "" {
		deviceInfo = nil
	}
	if ip == "" {
		ipAddr = nil
	}
	refreshToken := &model.RefreshToken{
		UserID:     user.ID,
		TokenHash:  refreshHash,
		DeviceInfo: deviceInfo,
		IPAddress:  ipAddr,
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.tokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshRaw,
		ExpiresIn:    expiresIn,
		User: dto.UserDTO{
			ID:          user.ID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarURL,
			Role:        user.Role,
			CreatedAt:   user.CreatedAt,
		},
	}, nil
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
