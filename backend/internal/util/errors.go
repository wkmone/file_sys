package util

import "errors"

var (
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidPassword    = errors.New("password must be at least 8 characters with letters and numbers")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountDeactivated = errors.New("account is deactivated")
	ErrEmailExists        = errors.New("email already registered")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrMissingToken       = errors.New("missing authorization token")
	ErrTokenRevoked       = errors.New("token has been revoked")

	ErrFileNotFound     = errors.New("file not found")
	ErrFileTooLarge     = errors.New("file size exceeds maximum limit")
	ErrInvalidFileType  = errors.New("unsupported file type")
	ErrFileExists       = errors.New("file already exists")

	ErrFolderNotFound   = errors.New("folder not found")
	ErrFolderExists     = errors.New("folder already exists")
	ErrCannotDeleteRoot = errors.New("cannot delete root folder")

	ErrTeamNotFound  = errors.New("team not found")
	ErrTeamExists    = errors.New("team already exists")
	ErrNotTeamMember = errors.New("not a member of this team")

	ErrInsufficientPerms = errors.New("insufficient permissions")
	ErrAdminRequired     = errors.New("admin privileges required")

	ErrStorageError   = errors.New("storage operation failed")
	ErrDatabaseError  = errors.New("database operation failed")
)