package util

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AccessClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

type OOClaims struct {
	FileID      string `json:"file_id"`
	UserID      string `json:"user_id"`
	Mode        string `json:"mode"`
	Permissions struct {
		Edit     bool `json:"edit"`
		Comment  bool `json:"comment"`
		Download bool `json:"download"`
	} `json:"permissions"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID, email, role, secret string) (string, error) {
	claims := AccessClaims{
		Sub:   userID,
		Email: email,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateAccessToken(tokenStr, secret string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func GenerateOOToken(fileID, userID, mode, secret string, canEdit bool) (string, error) {
	claims := OOClaims{
		FileID: fileID,
		UserID: userID,
		Mode:   mode,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	claims.Permissions.Edit = canEdit
	claims.Permissions.Comment = true
	claims.Permissions.Download = true

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateOOToken(tokenStr, secret string) (*OOClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &OOClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*OOClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

func GenerateFileAccessToken(fileID, versionNumber int, secret string) (string, error) {
	// Short-lived token for OnlyOffice file access
	return "", nil // placeholder for future use
}
