package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"file_sys/backend/internal/dto"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OnlyOfficeService struct {
	fileService *FileService
	jwtSecret   string
	dsURL       string
	callbackURL string
	db          *pgxpool.Pool
}

func NewOnlyOfficeService(
	fileService *FileService,
	jwtSecret string,
	dsURL string,
	callbackURL string,
	db *pgxpool.Pool,
) *OnlyOfficeService {
	return &OnlyOfficeService{
		fileService: fileService,
		jwtSecret:   jwtSecret,
		dsURL:       dsURL,
		callbackURL: callbackURL,
		db:          db,
	}
}

func (s *OnlyOfficeService) GenerateEditorConfig(ctx context.Context, userID, userName, fileID, mode string) (*dto.EditorConfigResponse, error) {
	file, err := s.fileService.GetByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("get file: %w", err)
	}

	// Document key = SHA-256(fileID + version)
	docKeyRaw := fmt.Sprintf("%s_%d", file.ID, file.CurrentVersion)
	docKeyHash := sha256.Sum256([]byte(docKeyRaw))
	docKey := hex.EncodeToString(docKeyHash[:])[:20]

	canEdit := mode == "edit"

	// Map file extension to OnlyOffice documentType
	docType := "word"
	switch file.FileExt {
	case ".xlsx":
		docType = "cell"
	case ".pptx":
		docType = "slide"
	}

	config := map[string]interface{}{
		"documentType": docType,
		"type":         "desktop",
		"width":        "100%",
		"height":       "100%",
		"document": map[string]interface{}{
			"fileType": file.FileExt[1:],
			"key":      docKey,
			"title":    file.OriginalName,
			"url":      fmt.Sprintf("%s/api/v1/oo/file/%s", s.callbackURL, fileID),
			"permissions": map[string]interface{}{
				"edit":     canEdit,
				"download": true,
				"review":   true,
			},
		},
		"editorConfig": map[string]interface{}{
			"callbackUrl": fmt.Sprintf("%s/api/v1/oo/callback/%s", s.callbackURL, fileID),
			"user": map[string]string{
				"id":   userID,
				"name": userName,
			},
			"mode": mode,
			"lang": "zh-CN",
		},
	}

	// Serialize and sign entire config as JWT (required when OO DS has JWT enabled)
	configJSON, _ := json.Marshal(config)
	var claims jwt.MapClaims
	json.Unmarshal(configJSON, &claims)
	claims["exp"] = time.Now().Add(24 * time.Hour).Unix()
	claims["iat"] = time.Now().Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("sign config: %w", err)
	}

	return &dto.EditorConfigResponse{Token: signed, Config: config}, nil
}

func (s *OnlyOfficeService) HandleCallback(ctx context.Context, cb *dto.OnlyOfficeCallback, fileID string) error {
	switch cb.Status {
	case 1: // editing started
		// Record editing session
		s.recordSession(ctx, fileID, cb)

	case 2, 6: // save / force-save
		if cb.URL != "" {
			resp, err := http.Get(cb.URL)
			if err != nil {
				return fmt.Errorf("download from OO: %w", err)
			}
			defer resp.Body.Close()

			createdBy := ""
			if len(cb.Users) > 0 {
				createdBy = cb.Users[0]
			}

			_, err = s.fileService.CreateNewVersion(ctx, fileID, resp.Body, createdBy, "OnlyOffice save")
			if err != nil {
				return fmt.Errorf("create version: %w", err)
			}
		}

	case 4: // closed without changes
		// No action
	}

	return nil
}

func (s *OnlyOfficeService) recordSession(ctx context.Context, fileID string, cb *dto.OnlyOfficeCallback) {
	// Store active editing session
	for _, uid := range cb.Users {
		s.db.Exec(ctx,
			`INSERT INTO onlyoffice_sessions (file_id, user_id, document_key, mode, status)
			 VALUES ($1, $2, $3, 'edit', 'active')
			 ON CONFLICT DO NOTHING`,
			fileID, uid, cb.Key)
	}
}

func (s *OnlyOfficeService) GetFileStream(ctx context.Context, fileID string) (io.ReadCloser, string, int64, error) {
	file, err := s.fileService.GetByID(ctx, fileID)
	if err != nil {
		return nil, "", 0, err
	}

	reader, info, err := s.fileService.RetrieveStorage(ctx, file.StorageKey)
	if err != nil {
		return nil, "", 0, err
	}

	ext := filepath.Ext(file.OriginalName)
	mimeType := detectMimeType_oo(ext)
	return reader, mimeType, info.Size, nil
}

func (s *OnlyOfficeService) GetJWTSecret() string {
	return s.jwtSecret
}

func (s *OnlyOfficeService) LookupFileIDByDocumentKey(ctx context.Context, docKey string) string {
	var fileID string
	err := s.db.QueryRow(ctx,
		`SELECT file_id FROM onlyoffice_sessions WHERE document_key = $1 ORDER BY created_at DESC LIMIT 1`,
		docKey).Scan(&fileID)
	if err != nil {
		return ""
	}
	return fileID
}

func detectMimeType_oo(ext string) string {
	mimeTypes := map[string]string{
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".pdf":  "application/pdf",
		".txt":  "text/plain",
	}
	if m, ok := mimeTypes[ext]; ok {
		return m
	}
	return "application/octet-stream"
}
