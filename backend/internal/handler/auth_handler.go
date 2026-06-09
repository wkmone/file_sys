package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"file_sys/backend/internal/dto"
	"file_sys/backend/internal/middleware"
	"file_sys/backend/internal/repository"
	"file_sys/backend/internal/service"
	"file_sys/backend/internal/util"
)

type AuthHandler struct {
	authService *service.AuthService
	userRepo    *repository.UserRepo
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) SetUserRepo(repo *repository.UserRepo) {
	h.userRepo = repo
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		util.Error(c, 409, 40901, err.Error())
		return
	}

	setRefreshCookie(c, resp.RefreshToken)
	resp.RefreshToken = ""
	util.Created(c, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}

	ip := c.ClientIP()
	device := c.GetHeader("User-Agent")
	resp, err := h.authService.Login(c.Request.Context(), &req, ip, device)
	if err != nil {
		util.Unauthorized(c, err.Error())
		return
	}

	setRefreshCookie(c, resp.RefreshToken)
	resp.RefreshToken = ""
	util.Success(c, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	rawToken, err := c.Cookie("refresh_token")
	if err != nil {
		util.Unauthorized(c, "missing refresh token")
		return
	}

	resp, err := h.authService.Refresh(c.Request.Context(), rawToken)
	if err != nil {
		util.Unauthorized(c, err.Error())
		return
	}

	setRefreshCookie(c, resp.RefreshToken)
	resp.RefreshToken = ""
	util.Success(c, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	rawToken, err := c.Cookie("refresh_token")
	if err == nil {
		h.authService.Logout(c.Request.Context(), rawToken)
	}
	c.SetCookie("refresh_token", "", -1, "/api/v1/auth", "", true, true)
	util.Success(c, nil)
}

func (h *AuthHandler) LogoutAll(c *gin.Context) {
	if err := h.authService.LogoutAll(c.Request.Context(), middleware.GetUserID(c)); err != nil {
		util.InternalError(c, "退出所有设备失败")
		return
	}
	c.SetCookie("refresh_token", "", -1, "/api/v1/auth", "", true, true)
	util.Success(c, nil)
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, err := h.authService.GetCurrentUser(c.Request.Context(), middleware.GetUserID(c))
	if err != nil {
		util.NotFound(c, "user not found")
		return
	}
	util.Success(c, dto.UserDTO{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Role:        user.Role,
		CreatedAt:   user.CreatedAt,
	})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), middleware.GetUserID(c), req.OldPassword, req.NewPassword); err != nil {
		util.Error(c, 400, 40002, err.Error())
		return
	}
	util.Success(c, nil)
}

func (h *AuthHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if h.userRepo == nil {
		util.Success(c, nil)
		return
	}
	users, total, err := h.userRepo.FindAll(c.Request.Context(), page, pageSize)
	if err != nil {
		util.InternalError(c, "failed to fetch users")
		return
	}
	util.Success(c, dto.PaginatedResponse{
		Items:      users,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	})
}

func (h *AuthHandler) GetUser(c *gin.Context) {
	if h.userRepo == nil {
		util.Success(c, nil)
		return
	}
	user, err := h.userRepo.FindByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		util.NotFound(c, "user not found")
		return
	}
	util.Success(c, dto.UserDTO{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Role:        user.Role,
		CreatedAt:   user.CreatedAt,
	})
}

func setRefreshCookie(c *gin.Context, token string) {
	c.SetCookie("refresh_token", token, 7*24*3600, "/api/v1/auth", "", true, true)
}
