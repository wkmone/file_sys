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
	fileService              *FileService
	jwtSecret                string
	dsURL                    string
	callbackURL              string
	db                       *pgxpool.Pool
	theme                    string
	jwtExpireHours           int
	docCacheEnabled          bool
	largeFileThresholdMB     int64
}

func NewOnlyOfficeService(
	fileService *FileService,
	jwtSecret string,
	dsURL string,
	callbackURL string,
	db *pgxpool.Pool,
	theme string,
	jwtExpireHours int,
	docCacheEnabled bool,
	largeFileThresholdMB int64,
) *OnlyOfficeService {
	return &OnlyOfficeService{
		fileService:              fileService,
		jwtSecret:                jwtSecret,
		dsURL:                    dsURL,
		callbackURL:              callbackURL,
		db:                       db,
		theme:                    theme,
		jwtExpireHours:           jwtExpireHours,
		docCacheEnabled:          docCacheEnabled,
		largeFileThresholdMB:     largeFileThresholdMB,
	}
}

func (s *OnlyOfficeService) GenerateEditorConfig(ctx context.Context, userID, userName, fileID, mode string) (*dto.EditorConfigResponse, error) {
	file, err := s.fileService.GetByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("get file: %w", err)
	}

	// Look up display name from DB
	displayName := userName
	if s.db != nil && userID != "" {
		var dn string
		if err := s.db.QueryRow(ctx, `SELECT display_name FROM users WHERE id = $1`, userID).Scan(&dn); err == nil {
			displayName = dn
		}
	}

	docKeyRaw := fmt.Sprintf("%s_%d", file.ID, file.CurrentVersion)
	docKeyHash := sha256.Sum256([]byte(docKeyRaw))
	docKey := hex.EncodeToString(docKeyHash[:])[:20]

	docType := detectDocumentType_oo(file.FileExt)

	activeUsers := s.getActiveUsers(ctx, fileID)

	config := map[string]interface{}{
		"documentType": docType,
		"type":         "desktop",
		"width":        "100%",
		"height":       "100%",
		"document": map[string]interface{}{
			"fileType":    file.FileExt[1:],
			"key":         docKey,
			"title":       file.OriginalName,
			"url":         fmt.Sprintf("%s/api/v1/oo/file/%s", s.callbackURL, fileID),
			"permissions": mapPermissions(mode),
		},
		"editorConfig": map[string]interface{}{
			"callbackUrl":  fmt.Sprintf("%s/api/v1/oo/callback/%s", s.callbackURL, fileID),
			"user": map[string]string{
				"id":   userID,
				"name": displayName,
			},
			"users":         activeUsers,
			"mode":          mode,
			"lang":          "zh-CN",
			"customization": buildCustomization(s.theme),
		},
	}

	configJSON, _ := json.Marshal(config)
	var claims jwt.MapClaims
	json.Unmarshal(configJSON, &claims)
	
	// 使用配置的 JWT 过期时间（小时）
	expireHours := s.jwtExpireHours
	if expireHours <= 0 {
		expireHours = 24
	}
	claims["exp"] = time.Now().Add(time.Duration(expireHours) * time.Hour).Unix()
	claims["iat"] = time.Now().Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("sign config: %w", err)
	}

	return &dto.EditorConfigResponse{Token: signed, Config: config}, nil
}

func detectDocumentType_oo(ext string) string {
	wordExts := map[string]bool{
		".docx": true, ".doc": true, ".odt": true, ".rtf": true,
		".txt": true, ".html": true, ".htm": true, ".mht": true,
		".epub": true, ".fb2": true, ".dotx": true, ".ott": true,
		".docxf": true, ".oform": true,
		".pdf": true, ".djvu": true, ".xps": true, ".oxps": true,
	}
	cellExts := map[string]bool{
		".xlsx": true, ".xls": true, ".ods": true, ".csv": true,
		".xltx": true, ".ots": true, ".fods": true,
	}
	slideExts := map[string]bool{
		".pptx": true, ".ppt": true, ".odp": true, ".ppsx": true,
		".pps": true, ".potx": true, ".otp": true,
	}
	if wordExts[ext] {
		return "word"
	}
	if cellExts[ext] {
		return "cell"
	}
	if slideExts[ext] {
		return "slide"
	}
	return "word"
}

func mapPermissions(mode string) map[string]interface{} {
	perms := map[string]interface{}{
		"edit": false, "comment": false, "review": false,
		"fillForms": false, "modifyFilter": false,
		"modifyContentControl": false, "copy": true,
		"download": true, "print": true,
	}
	switch mode {
	case "edit":
		perms["edit"] = true
		perms["comment"] = true
		perms["review"] = true
		perms["fillForms"] = true
		perms["modifyFilter"] = true
		perms["modifyContentControl"] = true
	case "comment":
		perms["comment"] = true
	case "review":
		perms["comment"] = true
		perms["review"] = true
	case "fillForms":
		perms["fillForms"] = true
	}
	return perms
}

func buildCustomization(theme string) map[string]interface{} {
	// Use default theme if not specified
	if theme == "" {
		theme = "theme-light"
	}
	return map[string]interface{}{
		"autosave":            true,
		"chat":                false,
		"comments":            true,
		"compactHeader":       true,
		"compactToolbar":      false,
		"forcesave":           true,
		"help":                false,
		"hideRightMenu":       false,
		"hideRulers":          false,
		"spellcheck":          true,
		"uiTheme":             theme,
		"toolbarHideFileName": false,
		"zoom":                100,
		"macros":              false,
		"plugins":             false,
	}
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
	for _, uid := range cb.Users {
		s.db.Exec(ctx,
			`INSERT INTO onlyoffice_sessions (file_id, user_id, document_key, mode, status)
			 VALUES ($1, $2, $3, 'edit', 'active')
			 ON CONFLICT DO NOTHING`,
			fileID, uid, cb.Key)
	}
}

func (s *OnlyOfficeService) getActiveUsers(ctx context.Context, fileID string) []map[string]string {
	if s.db == nil {
		return nil
	}
	rows, err := s.db.Query(ctx,
		`SELECT DISTINCT os.user_id, u.display_name
		 FROM onlyoffice_sessions os
		 JOIN users u ON os.user_id = u.id
		 WHERE os.file_id = $1 AND os.status = 'active'
		 ORDER BY os.created_at DESC LIMIT 20`, fileID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var users []map[string]string
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}
		users = append(users, map[string]string{"id": id, "name": name})
	}
	return users
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
		".docx":  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".doc":   "application/msword",
		".odt":   "application/vnd.oasis.opendocument.text",
		".rtf":   "application/rtf",
		".txt":   "text/plain",
		".html":  "text/html",
		".htm":   "text/html",
		".mht":   "multipart/related",
		".epub":  "application/epub+zip",
		".fb2":   "application/x-fictionbook+xml",
		".dotx":  "application/vnd.openxmlformats-officedocument.wordprocessingml.template",
		".ott":   "application/vnd.oasis.opendocument.text-template",
		".docxf": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".oform": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xlsx":  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".xls":   "application/vnd.ms-excel",
		".ods":   "application/vnd.oasis.opendocument.spreadsheet",
		".csv":   "text/csv",
		".xltx":  "application/vnd.openxmlformats-officedocument.spreadsheetml.template",
		".ots":   "application/vnd.oasis.opendocument.spreadsheet-template",
		".fods":  "application/vnd.oasis.opendocument.spreadsheet-flat-xml",
		".pptx":  "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".ppt":   "application/vnd.ms-powerpoint",
		".odp":   "application/vnd.oasis.opendocument.presentation",
		".ppsx":  "application/vnd.openxmlformats-officedocument.presentationml.slideshow",
		".pps":   "application/vnd.ms-powerpoint",
		".potx":  "application/vnd.openxmlformats-officedocument.presentationml.template",
		".otp":   "application/vnd.oasis.opendocument.presentation-template",
		".pdf":   "application/pdf",
		".djvu":  "image/vnd.djvu",
		".xps":   "application/vnd.ms-xpsdocument",
		".oxps":  "application/oxps",
	}
	if m, ok := mimeTypes[ext]; ok {
		return m
	}
	return "application/octet-stream"
}
