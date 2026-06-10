package service

import (
	"context"
	"errors"

	"file_sys/backend/internal/dto"
	"file_sys/backend/internal/model"
	"file_sys/backend/internal/repository"
)

type TeamService struct {
	teamRepo *repository.TeamRepo
	userRepo *repository.UserRepo
}

func NewTeamService(teamRepo *repository.TeamRepo, userRepo *repository.UserRepo) *TeamService {
	return &TeamService{teamRepo: teamRepo, userRepo: userRepo}
}

func (s *TeamService) Create(ctx context.Context, req *dto.CreateTeamRequest, ownerID string) (*model.Team, error) {
	team := &model.Team{
		Name:        req.Name,
		Description: &req.Description,
		OwnerID:     ownerID,
	}
	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, err
	}

	// Owner is automatically a member with "owner" role
	if err := s.teamRepo.AddMember(ctx, team.ID, ownerID, "owner"); err != nil {
		return nil, err
	}

	return team, nil
}

func (s *TeamService) GetByID(ctx context.Context, id string) (*model.Team, error) {
	return s.teamRepo.FindByID(ctx, id)
}

func (s *TeamService) ListByUser(ctx context.Context, userID string) ([]model.Team, error) {
	return s.teamRepo.FindByUser(ctx, userID)
}

func (s *TeamService) Update(ctx context.Context, id string, req *dto.UpdateTeamRequest) error {
	return s.teamRepo.Update(ctx, id, req.Name, req.Description)
}

func (s *TeamService) Delete(ctx context.Context, id, userID string) error {
	team, err := s.teamRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if team.OwnerID != userID {
		return errors.New("only team owner can delete the team")
	}
	return s.teamRepo.Delete(ctx, id)
}

func (s *TeamService) ListMembers(ctx context.Context, teamID string) ([]model.TeamMember, error) {
	return s.teamRepo.ListMembers(ctx, teamID)
}

func (s *TeamService) AddMember(ctx context.Context, teamID, addedBy string, req *dto.AddMemberRequest) error {
	// Check that adder is admin or owner
	role, err := s.teamRepo.GetMemberRole(ctx, teamID, addedBy)
	if err != nil || (role != "owner" && role != "admin") {
		return errors.New("insufficient permissions")
	}

	// Verify user exists
	_, err = s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	return s.teamRepo.AddMember(ctx, teamID, req.UserID, req.Role)
}

func (s *TeamService) RemoveMember(ctx context.Context, teamID, removedBy, targetUserID string) error {
	role, err := s.teamRepo.GetMemberRole(ctx, teamID, removedBy)
	if err != nil || (role != "owner" && role != "admin") {
		return errors.New("insufficient permissions")
	}

	// Cannot remove the owner
	targetRole, _ := s.teamRepo.GetMemberRole(ctx, teamID, targetUserID)
	if targetRole == "owner" {
		return errors.New("cannot remove the team owner")
	}

	return s.teamRepo.RemoveMember(ctx, teamID, targetUserID)
}

func (s *TeamService) UpdateMemberRole(ctx context.Context, teamID, updatedBy, targetUserID, newRole string) error {
	role, err := s.teamRepo.GetMemberRole(ctx, teamID, updatedBy)
	if err != nil || role != "owner" {
		return errors.New("only team owner can change roles")
	}
	return s.teamRepo.UpdateMemberRole(ctx, teamID, targetUserID, newRole)
}

func (s *TeamService) ListAllTeams(ctx context.Context) ([]model.Team, error) {
	return s.teamRepo.FindAll(ctx)
}

func (s *TeamService) RequestJoin(ctx context.Context, teamID, userID string) error {
	isMember, err := s.teamRepo.IsMember(ctx, teamID, userID)
	if err != nil {
		return err
	}
	if isMember {
		return errors.New("already a member of this team")
	}
	return s.teamRepo.AddMember(ctx, teamID, userID, "member")
}

func (s *TeamService) ListJoinRequests(ctx context.Context, teamID, userID string) ([]model.JoinRequest, error) {
	role, err := s.teamRepo.GetMemberRole(ctx, teamID, userID)
	if err != nil || (role != "owner" && role != "admin") {
		return nil, errors.New("insufficient permissions")
	}
	return s.teamRepo.ListJoinRequests(ctx, teamID)
}

func (s *TeamService) HandleJoinRequest(ctx context.Context, requestID, status, handledBy string) error {
	jr, err := s.teamRepo.GetJoinRequest(ctx, requestID)
	if err != nil {
		return err
	}

	role, err := s.teamRepo.GetMemberRole(ctx, jr.TeamID, handledBy)
	if err != nil || (role != "owner" && role != "admin") {
		return errors.New("insufficient permissions")
	}

	if err := s.teamRepo.HandleJoinRequest(ctx, requestID, status); err != nil {
		return err
	}

	if status == "approved" {
		_ = s.teamRepo.AddMember(ctx, jr.TeamID, jr.UserID, "member")
	}

	return nil
}

func (s *TeamService) GetPendingRequest(ctx context.Context, teamID, userID string) (*model.JoinRequest, error) {
	return s.teamRepo.GetPendingRequest(ctx, teamID, userID)
}
