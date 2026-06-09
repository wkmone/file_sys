package handler

import (
	"file_sys/backend/internal/dto"
	"file_sys/backend/internal/middleware"
	"file_sys/backend/internal/service"
	"file_sys/backend/internal/util"

	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	teamService *service.TeamService
}

func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

func (h *TeamHandler) List(c *gin.Context) {
	teams, err := h.teamService.ListByUser(c.Request.Context(), middleware.GetUserID(c))
	if err != nil {
		util.InternalError(c, "failed to list teams")
		return
	}
	util.Success(c, teams)
}

func (h *TeamHandler) Create(c *gin.Context) {
	var req dto.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}

	team, err := h.teamService.Create(c.Request.Context(), &req, middleware.GetUserID(c))
	if err != nil {
		util.InternalError(c, "create team failed")
		return
	}
	util.Created(c, team)
}

func (h *TeamHandler) Get(c *gin.Context) {
	team, err := h.teamService.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		util.NotFound(c, "team not found")
		return
	}
	util.Success(c, team)
}

func (h *TeamHandler) Update(c *gin.Context) {
	var req dto.UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}
	if err := h.teamService.Update(c.Request.Context(), c.Param("id"), &req); err != nil {
		util.InternalError(c, "update failed")
		return
	}
	util.Success(c, nil)
}

func (h *TeamHandler) Delete(c *gin.Context) {
	if err := h.teamService.Delete(c.Request.Context(), c.Param("id"), middleware.GetUserID(c)); err != nil {
		util.Error(c, 403, 40302, err.Error())
		return
	}
	util.Success(c, nil)
}

func (h *TeamHandler) Members(c *gin.Context) {
	members, err := h.teamService.ListMembers(c.Request.Context(), c.Param("id"))
	if err != nil {
		util.InternalError(c, "failed to list members")
		return
	}
	util.Success(c, members)
}

func (h *TeamHandler) AddMember(c *gin.Context) {
	var req dto.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}
	if err := h.teamService.AddMember(c.Request.Context(), c.Param("id"), middleware.GetUserID(c), &req); err != nil {
		util.Error(c, 403, 40303, err.Error())
		return
	}
	util.Success(c, nil)
}

func (h *TeamHandler) UpdateMember(c *gin.Context) {
	var req dto.UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}
	if err := h.teamService.UpdateMemberRole(c.Request.Context(), c.Param("id"), middleware.GetUserID(c), c.Param("userId"), req.Role); err != nil {
		util.Error(c, 403, 40304, err.Error())
		return
	}
	util.Success(c, nil)
}

func (h *TeamHandler) RemoveMember(c *gin.Context) {
	if err := h.teamService.RemoveMember(c.Request.Context(), c.Param("id"), middleware.GetUserID(c), c.Param("userId")); err != nil {
		util.Error(c, 403, 40305, err.Error())
		return
	}
	util.Success(c, nil)
}

func (h *TeamHandler) Discover(c *gin.Context) {
	teams, err := h.teamService.ListAllTeams(c.Request.Context())
	if err != nil {
		util.InternalError(c, "failed to list teams")
		return
	}
	util.Success(c, teams)
}

func (h *TeamHandler) RequestJoin(c *gin.Context) {
	if err := h.teamService.RequestJoin(c.Request.Context(), c.Param("id"), middleware.GetUserID(c)); err != nil {
		util.Error(c, 400, 40001, err.Error())
		return
	}
	util.Success(c, nil)
}

func (h *TeamHandler) PendingRequest(c *gin.Context) {
	jr, err := h.teamService.GetPendingRequest(c.Request.Context(), c.Param("id"), middleware.GetUserID(c))
	if err != nil {
		util.Success(c, nil)
		return
	}
	util.Success(c, jr)
}

func (h *TeamHandler) ListJoinRequests(c *gin.Context) {
	requests, err := h.teamService.ListJoinRequests(c.Request.Context(), c.Param("id"), middleware.GetUserID(c))
	if err != nil {
		util.Error(c, 403, 40306, err.Error())
		return
	}
	util.Success(c, requests)
}

func (h *TeamHandler) HandleJoinRequest(c *gin.Context) {
	var req dto.HandleJoinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}
	if err := h.teamService.HandleJoinRequest(c.Request.Context(), c.Param("requestId"), req.Status, middleware.GetUserID(c)); err != nil {
		util.Error(c, 403, 40307, err.Error())
		return
	}
	util.Success(c, nil)
}
