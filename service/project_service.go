package service

import (
	"context"
	"errors"

	"gotask/dto"
	"gotask/entity"
	"gotask/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrProjectNotFound      = errors.New("project tidak ditemukan")
	ErrUnauthorizedProject  = errors.New("Anda tidak memiliki izin untuk project ini")
	ErrMemberAlreadyExists  = errors.New("pengguna sudah menjadi member project")
	ErrMemberNotFound       = errors.New("member tidak ditemukan di project")
	ErrInvalidProjectID     = errors.New("ID project tidak valid")
	ErrInvalidUserID        = errors.New("ID user tidak valid")
)

type ProjectService interface {
	Create(ctx context.Context, userID string, req dto.CreateProjectRequest) (*dto.ProjectResponse, error)
	GetAll(ctx context.Context, userID string, role entity.Role) ([]dto.ProjectResponse, error)
	GetByID(ctx context.Context, projectID string, userID string, role entity.Role) (*dto.ProjectResponse, error)
	Update(ctx context.Context, projectID string, userID string, role entity.Role, req dto.UpdateProjectRequest) (*dto.ProjectResponse, error)
	Delete(ctx context.Context, projectID string, userID string, role entity.Role) error
	AddMember(ctx context.Context, projectID string, currentUserID string, role entity.Role, req dto.AddMemberRequest) (*dto.ProjectMemberResponse, error)
	GetMembers(ctx context.Context, projectID string, currentUserID string, role entity.Role) ([]dto.ProjectMemberResponse, error)
	RemoveMember(ctx context.Context, projectID string, targetUserID string, currentUserID string, role entity.Role) error
}

type projectService struct {
	projectRepo repository.ProjectRepository
	memberRepo  repository.ProjectMemberRepository
	userRepo    repository.UserRepository
	taskRepo    repository.TaskRepository
}

func NewProjectService(
	projectRepo repository.ProjectRepository,
	memberRepo repository.ProjectMemberRepository,
	userRepo repository.UserRepository,
	taskRepo repository.TaskRepository,
) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
		userRepo:    userRepo,
		taskRepo:    taskRepo,
	}
}

func (s *projectService) Create(ctx context.Context, userID string, req dto.CreateProjectRequest) (*dto.ProjectResponse, error) {
	creatorOID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	project := entity.Project{
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   creatorOID,
	}

	if err := s.projectRepo.Create(ctx, &project); err != nil {
		return nil, err
	}

	// Otomatis daftarkan creator sebagai member project
	_ = s.memberRepo.AddMember(ctx, &entity.ProjectMember{
		ProjectID: project.ID,
		UserID:    creatorOID,
	})

	res := formatProjectResponse(&project)
	return &res, nil
}

func (s *projectService) GetAll(ctx context.Context, userID string, role entity.Role) ([]dto.ProjectResponse, error) {
	if role == entity.RoleAdmin {
		projects, err := s.projectRepo.FindAll(ctx)
		if err != nil {
			return nil, err
		}
		responses := make([]dto.ProjectResponse, 0, len(projects))
		for i := range projects {
			responses = append(responses, formatProjectResponse(&projects[i]))
		}
		return responses, nil
	}

	// Untuk manager/member, cari project yang dibuat ATAU yang berstatus member
	createdProjects, err := s.projectRepo.FindByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	memberRecords, err := s.memberRepo.FindProjectsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	projectMap := make(map[string]entity.Project)
	for _, p := range createdProjects {
		projectMap[p.ID.Hex()] = p
	}

	memberProjectIDs := make([]primitive.ObjectID, 0, len(memberRecords))
	for _, m := range memberRecords {
		if _, exists := projectMap[m.ProjectID.Hex()]; !exists {
			memberProjectIDs = append(memberProjectIDs, m.ProjectID)
		}
	}

	if len(memberProjectIDs) > 0 {
		memberProjects, err := s.projectRepo.FindByIDs(ctx, memberProjectIDs)
		if err == nil {
			for _, p := range memberProjects {
				projectMap[p.ID.Hex()] = p
			}
		}
	}

	responses := make([]dto.ProjectResponse, 0, len(projectMap))
	for _, p := range projectMap {
		responses = append(responses, formatProjectResponse(&p))
	}
	return responses, nil
}

func (s *projectService) GetByID(ctx context.Context, projectID string, userID string, role entity.Role) (*dto.ProjectResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}

	if role != entity.RoleAdmin && project.CreatedBy.Hex() != userID {
		// Periksa apakah user adalah member
		member, err := s.memberRepo.FindByProjectAndUser(ctx, projectID, userID)
		if err != nil || member == nil {
			return nil, ErrUnauthorizedProject
		}
	}

	res := formatProjectResponse(project)
	return &res, nil
}

func (s *projectService) Update(ctx context.Context, projectID string, userID string, role entity.Role, req dto.UpdateProjectRequest) (*dto.ProjectResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}

	if role != entity.RoleAdmin && project.CreatedBy.Hex() != userID {
		return nil, ErrUnauthorizedProject
	}

	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}

	if err := s.projectRepo.Update(ctx, projectID, project); err != nil {
		return nil, err
	}

	res := formatProjectResponse(project)
	return &res, nil
}

func (s *projectService) Delete(ctx context.Context, projectID string, userID string, role entity.Role) error {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return err
	}
	if project == nil {
		return ErrProjectNotFound
	}

	if role != entity.RoleAdmin && project.CreatedBy.Hex() != userID {
		return ErrUnauthorizedProject
	}

	// Cascade delete tasks dan project members
	_ = s.taskRepo.DeleteByProjectID(ctx, projectID)
	_ = s.memberRepo.DeleteByProjectID(ctx, projectID)

	return s.projectRepo.Delete(ctx, projectID)
}

func (s *projectService) AddMember(ctx context.Context, projectID string, currentUserID string, role entity.Role, req dto.AddMemberRequest) (*dto.ProjectMemberResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}

	if role != entity.RoleAdmin && project.CreatedBy.Hex() != currentUserID {
		return nil, ErrUnauthorizedProject
	}

	targetUser, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if targetUser == nil {
		return nil, ErrUserNotFound
	}

	existingMember, err := s.memberRepo.FindByProjectAndUser(ctx, projectID, req.UserID)
	if err != nil {
		return nil, err
	}
	if existingMember != nil {
		return nil, ErrMemberAlreadyExists
	}

	pOID, _ := primitive.ObjectIDFromHex(projectID)
	uOID, _ := primitive.ObjectIDFromHex(req.UserID)

	member := entity.ProjectMember{
		ProjectID: pOID,
		UserID:    uOID,
	}

	if err := s.memberRepo.AddMember(ctx, &member); err != nil {
		return nil, err
	}

	userRes := formatUserResponse(targetUser)
	return &dto.ProjectMemberResponse{
		ID:        member.ID.Hex(),
		ProjectID: projectID,
		UserID:    req.UserID,
		User:      &userRes,
		JoinedAt:  member.JoinedAt,
	}, nil
}

func (s *projectService) GetMembers(ctx context.Context, projectID string, currentUserID string, role entity.Role) ([]dto.ProjectMemberResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}

	if role != entity.RoleAdmin && project.CreatedBy.Hex() != currentUserID {
		member, err := s.memberRepo.FindByProjectAndUser(ctx, projectID, currentUserID)
		if err != nil || member == nil {
			return nil, ErrUnauthorizedProject
		}
	}

	members, err := s.memberRepo.FindMembersByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ProjectMemberResponse, 0, len(members))
	for _, m := range members {
		res := dto.ProjectMemberResponse{
			ID:        m.ID.Hex(),
			ProjectID: m.ProjectID.Hex(),
			UserID:    m.UserID.Hex(),
			JoinedAt:  m.JoinedAt,
		}
		if u, err := s.userRepo.FindByID(ctx, m.UserID.Hex()); err == nil && u != nil {
			uRes := formatUserResponse(u)
			res.User = &uRes
		}
		responses = append(responses, res)
	}

	return responses, nil
}

func (s *projectService) RemoveMember(ctx context.Context, projectID string, targetUserID string, currentUserID string, role entity.Role) error {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return err
	}
	if project == nil {
		return ErrProjectNotFound
	}

	if role != entity.RoleAdmin && project.CreatedBy.Hex() != currentUserID {
		return ErrUnauthorizedProject
	}

	member, err := s.memberRepo.FindByProjectAndUser(ctx, projectID, targetUserID)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrMemberNotFound
	}

	return s.memberRepo.RemoveMember(ctx, projectID, targetUserID)
}

func formatProjectResponse(p *entity.Project) dto.ProjectResponse {
	return dto.ProjectResponse{
		ID:          p.ID.Hex(),
		Name:        p.Name,
		Description: p.Description,
		CreatedBy:   p.CreatedBy.Hex(),
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
