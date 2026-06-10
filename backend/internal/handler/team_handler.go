package handler

import (
	"file_sys/backend/internal/dto"
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
	userID, _ := c.Get("user_id")
	teams, err := h.teamService.ListByUser(c.Request.Context(), userID.(string))
	if err != nil {
		util.DatabaseError(c)
		return
	}

	util.Success(c, teams)
}

func (h *TeamHandler) Discover(c *gin.Context) {
	teams, err := h.teamService.ListAllTeams(c.Request.Context())
	if err != nil {
		util.DatabaseError(c)
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

	userID, _ := c.Get("user_id")
	team, err := h.teamService.Create(c.Request.Context(), &req, userID.(string))
	if err != nil {
		util.Conflict(c, "team already exists")
		return
	}

	util.Created(c, team)
}

func (h *TeamHandler) Get(c *gin.Context) {
	team, err := h.teamService.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		util.TeamNotFound(c)
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

	err := h.teamService.Update(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		util.TeamNotFound(c)
		return
	}

	util.Success(c, nil)
}

func (h *TeamHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("user_id")
	err := h.teamService.Delete(c.Request.Context(), c.Param("id"), userID.(string))
	if err != nil {
		util.TeamNotFound(c)
		return
	}

	util.Success(c, nil)
}

func (h *TeamHandler) Members(c *gin.Context) {
	members, err := h.teamService.ListMembers(c.Request.Context(), c.Param("id"))
	if err != nil {
		util.TeamNotFound(c)
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

	userID, _ := c.Get("user_id")
	err := h.teamService.AddMember(c.Request.Context(), c.Param("id"), userID.(string), &req)
	if err != nil {
		util.Conflict(c, "member already exists")
		return
	}

	util.Created(c, nil)
}

func (h *TeamHandler) UpdateMember(c *gin.Context) {
	var req dto.UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.ValidationError(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	err := h.teamService.UpdateMemberRole(c.Request.Context(), c.Param("id"), userID.(string), c.Param("userId"), req.Role)
	if err != nil {
		util.TeamNotFound(c)
		return
	}

	util.Success(c, nil)
}

func (h *TeamHandler) RemoveMember(c *gin.Context) {
	userID, _ := c.Get("user_id")
	err := h.teamService.RemoveMember(c.Request.Context(), c.Param("id"), userID.(string), c.Param("userId"))
	if err != nil {
		util.TeamNotFound(c)
		return
	}

	util.Success(c, nil)
}

func (h *TeamHandler) RequestJoin(c *gin.Context) {
	userID, _ := c.Get("user_id")
	err := h.teamService.RequestJoin(c.Request.Context(), c.Param("id"), userID.(string))
	if err != nil {
		util.Conflict(c, "request already exists")
		return
	}

	util.Created(c, nil)
}

func (h *TeamHandler) PendingRequest(c *gin.Context) {
	userID, _ := c.Get("user_id")
	req, err := h.teamService.GetPendingRequest(c.Request.Context(), c.Param("id"), userID.(string))
	if err != nil {
		util.NotFound(c, "no pending request")
		return
	}

	util.Success(c, req)
}

func (h *TeamHandler) ListJoinRequests(c *gin.Context) {
	userID, _ := c.Get("user_id")
	requests, err := h.teamService.ListJoinRequests(c.Request.Context(), c.Param("id"), userID.(string))
	if err != nil {
		util.TeamNotFound(c)
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

	userID, _ := c.Get("user_id")
	err := h.teamService.HandleJoinRequest(c.Request.Context(), c.Param("requestId"), req.Status, userID.(string))
	if err != nil {
		util.NotFound(c, "request not found")
		return
	}

	util.Success(c, nil)
}
