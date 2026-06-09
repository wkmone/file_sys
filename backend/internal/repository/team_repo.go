package repository

import (
	"context"
	"time"

	"file_sys/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamRepo struct {
	db *pgxpool.Pool
}

func NewTeamRepo(db *pgxpool.Pool) *TeamRepo {
	return &TeamRepo{db: db}
}

func (r *TeamRepo) Create(ctx context.Context, team *model.Team) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO teams (name, description, owner_id) VALUES ($1, $2, $3)
		 RETURNING id, created_at, updated_at`,
		team.Name, team.Description, team.OwnerID,
	).Scan(&team.ID, &team.CreatedAt, &team.UpdatedAt)
}

func (r *TeamRepo) FindByID(ctx context.Context, id string) (*model.Team, error) {
	t := &model.Team{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, owner_id, created_at, updated_at
		 FROM teams WHERE id = $1`, id,
	).Scan(&t.ID, &t.Name, &t.Description, &t.OwnerID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TeamRepo) FindByUser(ctx context.Context, userID string) ([]model.Team, error) {
	rows, err := r.db.Query(ctx,
		`SELECT t.id, t.name, t.description, t.owner_id, t.created_at, t.updated_at
		 FROM teams t
		 INNER JOIN team_members tm ON t.id = tm.team_id
		 WHERE tm.user_id = $1
		 ORDER BY t.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []model.Team
	for rows.Next() {
		var t model.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.OwnerID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}

func (r *TeamRepo) Update(ctx context.Context, id string, name, description *string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE teams SET name = COALESCE($2, name), description = COALESCE($3, description), updated_at = $4
		 WHERE id = $1`, id, name, description, time.Now())
	return err
}

func (r *TeamRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM teams WHERE id = $1`, id)
	return err
}

func (r *TeamRepo) AddMember(ctx context.Context, teamID, userID, role string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3)
		 ON CONFLICT (team_id, user_id) DO UPDATE SET role = $3`,
		teamID, userID, role)
	return err
}

func (r *TeamRepo) RemoveMember(ctx context.Context, teamID, userID string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`, teamID, userID)
	return err
}

func (r *TeamRepo) ListMembers(ctx context.Context, teamID string) ([]model.TeamMember, error) {
	rows, err := r.db.Query(ctx,
		`SELECT tm.id, tm.team_id, tm.user_id, u.display_name, u.email, tm.role, tm.joined_at
		 FROM team_members tm
		 INNER JOIN users u ON u.id = tm.user_id
		 WHERE tm.team_id = $1
		 ORDER BY tm.role ASC, tm.joined_at ASC`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []model.TeamMember
	for rows.Next() {
		var m model.TeamMember
		if err := rows.Scan(&m.ID, &m.TeamID, &m.UserID, &m.DisplayName, &m.Email, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *TeamRepo) UpdateMemberRole(ctx context.Context, teamID, userID, role string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE team_members SET role = $3 WHERE team_id = $1 AND user_id = $2`,
		teamID, userID, role)
	return err
}

func (r *TeamRepo) GetMemberRole(ctx context.Context, teamID, userID string) (string, error) {
	var role string
	err := r.db.QueryRow(ctx,
		`SELECT role FROM team_members WHERE team_id = $1 AND user_id = $2`,
		teamID, userID).Scan(&role)
	if err != nil {
		return "", err
	}
	return role, nil
}

func (r *TeamRepo) FindAll(ctx context.Context) ([]model.Team, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, owner_id, created_at, updated_at
		 FROM teams ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []model.Team
	for rows.Next() {
		var t model.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.OwnerID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}

func (r *TeamRepo) IsMember(ctx context.Context, teamID, userID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2)`,
		teamID, userID).Scan(&exists)
	return exists, err
}

func (r *TeamRepo) CreateJoinRequest(ctx context.Context, teamID, userID string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO join_requests (team_id, user_id) VALUES ($1, $2)
		 ON CONFLICT (team_id, user_id) DO UPDATE SET status = 'pending', updated_at = now()`,
		teamID, userID)
	return err
}

func (r *TeamRepo) ListJoinRequests(ctx context.Context, teamID string) ([]model.JoinRequest, error) {
	rows, err := r.db.Query(ctx,
		`SELECT jr.id, jr.team_id, jr.user_id, jr.status, jr.created_at, jr.updated_at,
		        u.display_name, u.email
		 FROM join_requests jr
		 INNER JOIN users u ON u.id = jr.user_id
		 WHERE jr.team_id = $1
		 ORDER BY jr.created_at DESC`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []model.JoinRequest
	for rows.Next() {
		var r model.JoinRequest
		if err := rows.Scan(&r.ID, &r.TeamID, &r.UserID, &r.Status, &r.CreatedAt, &r.UpdatedAt,
			&r.DisplayName, &r.Email); err != nil {
			return nil, err
		}
		requests = append(requests, r)
	}
	return requests, nil
}

func (r *TeamRepo) HandleJoinRequest(ctx context.Context, requestID, status string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE join_requests SET status = $2, updated_at = now() WHERE id = $1`,
		requestID, status)
	return err
}

func (r *TeamRepo) GetJoinRequest(ctx context.Context, requestID string) (*model.JoinRequest, error) {
	var jr model.JoinRequest
	err := r.db.QueryRow(ctx,
		`SELECT id, team_id, user_id, status, created_at, updated_at
		 FROM join_requests WHERE id = $1`, requestID,
	).Scan(&jr.ID, &jr.TeamID, &jr.UserID, &jr.Status, &jr.CreatedAt, &jr.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &jr, nil
}

func (r *TeamRepo) GetPendingRequest(ctx context.Context, teamID, userID string) (*model.JoinRequest, error) {
	var jr model.JoinRequest
	err := r.db.QueryRow(ctx,
		`SELECT id, team_id, user_id, status, created_at, updated_at
		 FROM join_requests WHERE team_id = $1 AND user_id = $2 AND status = 'pending'`,
		teamID, userID,
	).Scan(&jr.ID, &jr.TeamID, &jr.UserID, &jr.Status, &jr.CreatedAt, &jr.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &jr, nil
}
