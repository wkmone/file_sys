package dto

import "time"

// Auth
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"display_name" binding:"required,min=1,max=128"`
}

type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	User         UserDTO   `json:"user"`
}

type UserDTO struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
	Role        string  `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// File
type UploadFileResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	OriginalName   string    `json:"original_name"`
	FolderID       *string   `json:"folder_id"`
	OwnerID        string    `json:"owner_id"`
	MimeType       string    `json:"mime_type"`
	FileSize       int64     `json:"file_size"`
	FileExt        string    `json:"file_ext"`
	CurrentVersion int       `json:"current_version"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UpdateFileRequest struct {
	Name     *string `json:"name"`
	FolderID *string `json:"folder_id"`
}

type CreateBlankFileRequest struct {
	Name     string  `json:"name" binding:"required,min=1,max=255"`
	FileExt  string  `json:"file_ext" binding:"required,oneof=.docx .xlsx .pptx"`
	FolderID *string `json:"folder_id"`
}

// Batch upload
type BatchUploadManifestEntry struct {
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	IsDirectory bool   `json:"is_directory"`
}

type BatchUploadFileResult struct {
	Path   string `json:"path"`
	FileID string `json:"file_id,omitempty"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type BatchUploadResponse struct {
	Total     int                     `json:"total"`
	Succeeded int                     `json:"succeeded"`
	Failed    int                     `json:"failed"`
	Results   []BatchUploadFileResult `json:"results"`
}

// Folder
type CreateFolderRequest struct {
	Name     string  `json:"name" binding:"required,min=1,max=255"`
	ParentID *string `json:"parent_id"`
	TeamID   *string `json:"team_id"`
}

type UpdateFolderRequest struct {
	Name     *string `json:"name"`
	ParentID *string `json:"parent_id"`
}

type FolderBreadcrumbItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type FolderDetailResponse struct {
	ID         string                  `json:"id"`
	Name       string                  `json:"name"`
	ParentID   *string                 `json:"parent_id"`
	OwnerID    string                  `json:"owner_id"`
	TeamID     *string                 `json:"team_id"`
	FolderPath string                  `json:"folder_path"`
	IsDeleted  bool                    `json:"is_deleted"`
	DeletedAt  *time.Time              `json:"deleted_at"`
	CreatedAt  time.Time               `json:"created_at"`
	UpdatedAt  time.Time               `json:"updated_at"`
	Breadcrumb []FolderBreadcrumbItem  `json:"breadcrumb"`
}

type ShareRequest struct {
	UserID     *string `json:"user_id"`
	TeamID     *string `json:"team_id"`
	Permission string  `json:"permission" binding:"required,oneof=read write admin"`
}

// Team
type CreateTeamRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=128"`
	Description string `json:"description"`
}

type UpdateTeamRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=admin member"`
}

type UpdateMemberRequest struct {
	Role string `json:"role" binding:"required,oneof=admin member"`
}

type HandleJoinRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
}

// OnlyOffice
type EditorConfigRequest struct {
	FileID string `json:"file_id" binding:"required"`
	Mode   string `json:"mode" binding:"required,oneof=edit view comment review fillForms"`
}

type EditorConfigResponse struct {
	Token  string                 `json:"token"`
	Config map[string]interface{} `json:"config,omitempty"`
}

type DocumentConfig struct {
	FileType    string              `json:"fileType"`
	Key         string              `json:"key"`
	Title       string              `json:"title"`
	URL         string              `json:"url"`
	Permissions DocumentPermissions `json:"permissions"`
}

type DocumentPermissions struct {
	Edit                 bool `json:"edit"`
	Comment              bool `json:"comment"`
	Review               bool `json:"review"`
	FillForms            bool `json:"fillForms"`
	ModifyFilter         bool `json:"modifyFilter"`
	ModifyContentControl bool `json:"modifyContentControl"`
	Copy                 bool `json:"copy"`
	Download             bool `json:"download"`
	Print                bool `json:"print"`
}

type EditorSettings struct {
	CallbackURL   string              `json:"callbackUrl"`
	User          EditorUser          `json:"user"`
	Mode          string              `json:"mode"`
	Lang          string              `json:"lang"`
	Customization EditorCustomization `json:"customization"`
}

type EditorUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type EditorCustomization struct {
	Autosave            bool   `json:"autosave"`
	Chat                bool   `json:"chat"`
	Comments            bool   `json:"comments"`
	CompactHeader       bool   `json:"compactHeader"`
	CompactToolbar      bool   `json:"compactToolbar"`
	Forcesave           bool   `json:"forcesave"`
	Help                bool   `json:"help"`
	HideRightMenu       bool   `json:"hideRightMenu"`
	HideRulers          bool   `json:"hideRulers"`
	Spellcheck          bool   `json:"spellcheck"`
	UiTheme             string `json:"uiTheme"`
	ToolbarHideFileName bool   `json:"toolbarHideFileName"`
	Zoom                int    `json:"zoom"`
	Macros              bool   `json:"macros"`
	Plugins             bool   `json:"plugins"`
}

type OnlyOfficeCallback struct {
	Key     string              `json:"key"`
	Status  int                 `json:"status"`
	URL     string              `json:"url"`
	Users   []string            `json:"users"`
	Actions []CallbackAction    `json:"actions"`
}

type CallbackAction struct {
	Type   int    `json:"type"`
	UserID string `json:"userid"`
}

// Pagination
type PaginationRequest struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=50"`
}

type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}
