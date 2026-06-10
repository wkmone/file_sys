package model

import "time"

type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	DisplayName  string     `json:"display_name"`
	AvatarURL    *string    `json:"avatar_url"`
	Role         string     `json:"role"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type Team struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TeamMember struct {
	ID          string    `json:"id"`
	TeamID      string    `json:"team_id"`
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

type Folder struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	ParentID  *string    `json:"parent_id"`
	OwnerID   string     `json:"owner_id"`
	TeamID    *string    `json:"team_id"`
	FolderPath string    `json:"folder_path"`
	IsDeleted bool       `json:"is_deleted"`
	DeletedAt *time.Time `json:"deleted_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type File struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	OriginalName   string     `json:"original_name"`
	FolderID       *string    `json:"folder_id"`
	OwnerID        string     `json:"owner_id"`
	OwnerName      string     `json:"owner_name"`
	TeamID         *string    `json:"team_id"`
	MimeType       string     `json:"mime_type"`
	FileSize       int64      `json:"file_size"`
	FileExt        string     `json:"file_ext"`
	StorageKey     string     `json:"-"`
	ContentHash    string     `json:"content_hash"`
	CurrentVersion int        `json:"current_version"`
	IsDeleted      bool       `json:"is_deleted"`
	DeletedAt      *time.Time `json:"deleted_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Permission     string     `json:"permission,omitempty"`
	SharedBy       string     `json:"shared_by,omitempty"`
}

type FileVersion struct {
	ID            string    `json:"id"`
	FileID        string    `json:"file_id"`
	VersionNumber int       `json:"version_number"`
	StorageKey    string    `json:"-"`
	FileSize      int64     `json:"file_size"`
	ContentHash   string    `json:"content_hash"`
	CreatedBy     *string   `json:"created_by"`
	ChangeNote    *string   `json:"change_note"`
	CreatedAt     time.Time `json:"created_at"`
}

type Permission struct {
	ID         string    `json:"id"`
	FolderID   *string   `json:"folder_id"`
	FileID     *string   `json:"file_id"`
	UserID     *string   `json:"user_id"`
	TeamID     *string   `json:"team_id"`
	Permission string    `json:"permission"`
	GrantedBy  string    `json:"granted_by"`
	CreatedAt  time.Time `json:"created_at"`
}

type RefreshToken struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	TokenHash  string    `json:"-"`
	DeviceInfo *string   `json:"device_info"`
	IPAddress  *string   `json:"ip_address"`
	ExpiresAt  time.Time `json:"expires_at"`
	Revoked    bool      `json:"revoked"`
	CreatedAt  time.Time `json:"created_at"`
}

type OnlyOfficeSession struct {
	ID          string    `json:"id"`
	FileID      string    `json:"file_id"`
	UserID      string    `json:"user_id"`
	DocumentKey string    `json:"document_key"`
	CallbackURL *string   `json:"callback_url"`
	Mode        string    `json:"mode"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type JoinRequest struct {
	ID        string    `json:"id"`
	TeamID    string    `json:"team_id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// Joined fields
	TeamName    string `json:"team_name,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Email       string `json:"email,omitempty"`
}
