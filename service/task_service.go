package service

import (
	"context"
	"errors"
	"strings"

	"gotask/dto"
	"gotask/entity"
	"gotask/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrTaskNotFound       = errors.New("task tidak ditemukan")
	ErrUnauthorizedTask   = errors.New("Anda tidak memiliki izin untuk task ini")
	ErrInvalidTaskStatus  = errors.New("status task tidak valid, pilih todo, in_progress, atau done")
	ErrInvalidPriority    = errors.New("priority task tidak valid, pilih low, medium, atau high")
	ErrAssigneeNotMember  = errors.New("pengguna yang ditugaskan bukan member dari project ini")
	ErrInvalidAssigneeID  = errors.New("ID pengguna yang ditugaskan tidak valid")
)

type TaskService interface {
	Create(ctx context.Context, projectID string, currentUserID string, role entity.Role, req dto.CreateTaskRequest) (*dto.TaskResponse, error)
	GetByProject(ctx context.Context, projectID string, currentUserID string, role entity.Role) ([]dto.TaskResponse, error)
	GetMyTasks(ctx context.Context, currentUserID string) ([]dto.TaskResponse, error)
	GetByID(ctx context.Context, taskID string, currentUserID string, role entity.Role) (*dto.TaskResponse, error)
	Update(ctx context.Context, taskID string, currentUserID string, role entity.Role, req dto.UpdateTaskRequest) (*dto.TaskResponse, error)
	UpdateStatus(ctx context.Context, taskID string, currentUserID string, role entity.Role, req dto.UpdateTaskStatusRequest) (*dto.TaskResponse, error)
	Delete(ctx context.Context, taskID string, currentUserID string, role entity.Role) error
	GetDashboard(ctx context.Context, currentUserID string, role entity.Role) (*dto.DashboardResponse, error)
}

type taskService struct {
	taskRepo    repository.TaskRepository
	projectRepo repository.ProjectRepository
	memberRepo  repository.ProjectMemberRepository
	userRepo    repository.UserRepository
}

func NewTaskService(
	taskRepo repository.TaskRepository,
	projectRepo repository.ProjectRepository,
	memberRepo repository.ProjectMemberRepository,
	userRepo repository.UserRepository,
) TaskService {
	return &taskService{
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
		userRepo:    userRepo,
	}
}

func (s *taskService) Create(ctx context.Context, projectID string, currentUserID string, role entity.Role, req dto.CreateTaskRequest) (*dto.TaskResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}

	if role != entity.RoleAdmin && project.CreatedBy.Hex() != currentUserID {
		return nil, ErrUnauthorizedTask
	}

	priority := entity.PriorityMedium
	if req.Priority != "" {
		p := entity.TaskPriority(strings.ToLower(strings.TrimSpace(req.Priority)))
		if p != entity.PriorityLow && p != entity.PriorityMedium && p != entity.PriorityHigh {
			return nil, ErrInvalidPriority
		}
		priority = p
	}

	var assignedOID *primitive.ObjectID
	if req.AssignedTo != nil && *req.AssignedTo != "" {
		oid, err := primitive.ObjectIDFromHex(*req.AssignedTo)
		if err != nil {
			return nil, ErrInvalidAssigneeID
		}
		// Validasi user exists
		targetUser, err := s.userRepo.FindByID(ctx, *req.AssignedTo)
		if err != nil || targetUser == nil {
			return nil, ErrUserNotFound
		}
		// Validasi user adalah member project atau pembuat
		if project.CreatedBy.Hex() != *req.AssignedTo {
			member, err := s.memberRepo.FindByProjectAndUser(ctx, projectID, *req.AssignedTo)
			if err != nil || member == nil {
				return nil, ErrAssigneeNotMember
			}
		}
		assignedOID = &oid
	}

	creatorOID, err := primitive.ObjectIDFromHex(currentUserID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	task := entity.Task{
		ProjectID:   project.ID,
		Title:       req.Title,
		Description: req.Description,
		Status:      entity.StatusTodo,
		Priority:    priority,
		AssignedTo:  assignedOID,
		Deadline:    req.Deadline,
		CreatedBy:   creatorOID,
	}

	if err := s.taskRepo.Create(ctx, &task); err != nil {
		return nil, err
	}

	res := formatTaskResponse(&task)
	return &res, nil
}

func (s *taskService) GetByProject(ctx context.Context, projectID string, currentUserID string, role entity.Role) ([]dto.TaskResponse, error) {
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
			return nil, ErrUnauthorizedTask
		}
	}

	tasks, err := s.taskRepo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.TaskResponse, 0, len(tasks))
	for i := range tasks {
		responses = append(responses, formatTaskResponse(&tasks[i]))
	}
	return responses, nil
}

func (s *taskService) GetMyTasks(ctx context.Context, currentUserID string) ([]dto.TaskResponse, error) {
	tasks, err := s.taskRepo.FindByAssignedTo(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.TaskResponse, 0, len(tasks))
	for i := range tasks {
		responses = append(responses, formatTaskResponse(&tasks[i]))
	}
	return responses, nil
}

func (s *taskService) GetByID(ctx context.Context, taskID string, currentUserID string, role entity.Role) (*dto.TaskResponse, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	if role != entity.RoleAdmin && task.CreatedBy.Hex() != currentUserID {
		if task.AssignedTo == nil || task.AssignedTo.Hex() != currentUserID {
			// Periksa juga membership di project
			member, err := s.memberRepo.FindByProjectAndUser(ctx, task.ProjectID.Hex(), currentUserID)
			if err != nil || member == nil {
				return nil, ErrUnauthorizedTask
			}
		}
	}

	res := formatTaskResponse(task)
	return &res, nil
}

func (s *taskService) Update(ctx context.Context, taskID string, currentUserID string, role entity.Role, req dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	// Hanya Admin atau pembuat task/project yang dapat mengedit task secara menyeluruh
	project, _ := s.projectRepo.FindByID(ctx, task.ProjectID.Hex())
	if role != entity.RoleAdmin && task.CreatedBy.Hex() != currentUserID {
		if project == nil || project.CreatedBy.Hex() != currentUserID {
			return nil, ErrUnauthorizedTask
		}
	}

	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Priority != "" {
		p := entity.TaskPriority(strings.ToLower(strings.TrimSpace(req.Priority)))
		if p != entity.PriorityLow && p != entity.PriorityMedium && p != entity.PriorityHigh {
			return nil, ErrInvalidPriority
		}
		task.Priority = p
	}
	if req.Status != "" {
		st := entity.TaskStatus(strings.ToLower(strings.TrimSpace(req.Status)))
		if st != entity.StatusTodo && st != entity.StatusInProgress && st != entity.StatusDone {
			return nil, ErrInvalidTaskStatus
		}
		task.Status = st
	}
	if req.AssignedTo != nil {
		if *req.AssignedTo == "" {
			task.AssignedTo = nil
		} else {
			oid, err := primitive.ObjectIDFromHex(*req.AssignedTo)
			if err != nil {
				return nil, ErrInvalidAssigneeID
			}
			task.AssignedTo = &oid
		}
	}
	if req.Deadline != nil {
		task.Deadline = req.Deadline
	}

	if err := s.taskRepo.Update(ctx, taskID, task); err != nil {
		return nil, err
	}

	res := formatTaskResponse(task)
	return &res, nil
}

func (s *taskService) UpdateStatus(ctx context.Context, taskID string, currentUserID string, role entity.Role, req dto.UpdateTaskStatusRequest) (*dto.TaskResponse, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	st := entity.TaskStatus(strings.ToLower(strings.TrimSpace(req.Status)))
	if st != entity.StatusTodo && st != entity.StatusInProgress && st != entity.StatusDone {
		return nil, ErrInvalidTaskStatus
	}

	// Status dapat diupdate oleh Admin, Creator, ATAU user yang ditugaskan (AssignedTo)
	isAssigned := task.AssignedTo != nil && task.AssignedTo.Hex() == currentUserID
	isCreator := task.CreatedBy.Hex() == currentUserID
	if role != entity.RoleAdmin && !isCreator && !isAssigned {
		project, _ := s.projectRepo.FindByID(ctx, task.ProjectID.Hex())
		if project == nil || project.CreatedBy.Hex() != currentUserID {
			return nil, ErrUnauthorizedTask
		}
	}

	if err := s.taskRepo.UpdateStatus(ctx, taskID, st); err != nil {
		return nil, err
	}

	task.Status = st
	res := formatTaskResponse(task)
	return &res, nil
}

func (s *taskService) Delete(ctx context.Context, taskID string, currentUserID string, role entity.Role) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}

	project, _ := s.projectRepo.FindByID(ctx, task.ProjectID.Hex())
	if role != entity.RoleAdmin && task.CreatedBy.Hex() != currentUserID {
		if project == nil || project.CreatedBy.Hex() != currentUserID {
			return ErrUnauthorizedTask
		}
	}

	return s.taskRepo.Delete(ctx, taskID)
}

func (s *taskService) GetDashboard(ctx context.Context, currentUserID string, role entity.Role) (*dto.DashboardResponse, error) {
	totalProjects, err := s.projectRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	totalTasks, err := s.taskRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	todoCount, _ := s.taskRepo.CountByStatus(ctx, entity.StatusTodo)
	inProgressCount, _ := s.taskRepo.CountByStatus(ctx, entity.StatusInProgress)
	doneCount, _ := s.taskRepo.CountByStatus(ctx, entity.StatusDone)

	return &dto.DashboardResponse{
		TotalProjects:   totalProjects,
		TotalTasks:      totalTasks,
		TasksTodo:       todoCount,
		TasksInProgress: inProgressCount,
		TasksDone:       doneCount,
	}, nil
}

func formatTaskResponse(t *entity.Task) dto.TaskResponse {
	var assignedTo *string
	if t.AssignedTo != nil {
		s := t.AssignedTo.Hex()
		assignedTo = &s
	}

	return dto.TaskResponse{
		ID:          t.ID.Hex(),
		ProjectID:   t.ProjectID.Hex(),
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.Status),
		Priority:    string(t.Priority),
		AssignedTo:  assignedTo,
		Deadline:    t.Deadline,
		CreatedBy:   t.CreatedBy.Hex(),
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
