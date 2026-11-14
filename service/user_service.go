package service

import (
	"context"
	"errors"
	"strings"

	"gotask/dto"
	"gotask/entity"
	"gotask/middleware"
	"gotask/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("pengguna tidak ditemukan")
	ErrEmailAlreadyExist = errors.New("email sudah terdaftar")
	ErrInvalidCredential = errors.New("email atau password salah")
	ErrInvalidRole       = errors.New("role tidak valid, pilih admin, manager, atau member")
)

type UserService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	GetAll(ctx context.Context) ([]dto.UserResponse, error)
	GetByID(ctx context.Context, id string) (*dto.UserResponse, error)
	Update(ctx context.Context, id string, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	Delete(ctx context.Context, id string) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyExist
	}

	role := entity.RoleMember
	if req.Role != "" {
		trimmedRole := entity.Role(strings.ToLower(strings.TrimSpace(req.Role)))
		if trimmedRole != entity.RoleAdmin && trimmedRole != entity.RoleManager && trimmedRole != entity.RoleMember {
			return nil, ErrInvalidRole
		}
		role = trimmedRole
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     role,
	}

	if err := s.userRepo.Create(ctx, &user); err != nil {
		return nil, err
	}

	res := formatUserResponse(&user)
	return &res, nil
}

func (s *userService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredential
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredential
	}

	token, err := middleware.GenerateToken(user.ID.Hex(), user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		Token: token,
		User:  formatUserResponse(user),
	}, nil
}

func (s *userService) GetAll(ctx context.Context) ([]dto.UserResponse, error) {
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.UserResponse, 0, len(users))
	for i := range users {
		responses = append(responses, formatUserResponse(&users[i]))
	}
	return responses, nil
}

func (s *userService) GetByID(ctx context.Context, id string) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	res := formatUserResponse(user)
	return &res, nil
}

func (s *userService) Update(ctx context.Context, id string, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		email := strings.TrimSpace(strings.ToLower(req.Email))
		if email != user.Email {
			checkUser, err := s.userRepo.FindByEmail(ctx, email)
			if err != nil {
				return nil, err
			}
			if checkUser != nil && checkUser.ID != user.ID {
				return nil, ErrEmailAlreadyExist
			}
			user.Email = email
		}
	}
	if req.Role != "" {
		role := entity.Role(strings.ToLower(strings.TrimSpace(req.Role)))
		if role != entity.RoleAdmin && role != entity.RoleManager && role != entity.RoleMember {
			return nil, ErrInvalidRole
		}
		user.Role = role
	}
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashed)
	}

	if err := s.userRepo.Update(ctx, id, user); err != nil {
		return nil, err
	}

	res := formatUserResponse(user)
	return &res, nil
}

func (s *userService) Delete(ctx context.Context, id string) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	return s.userRepo.Delete(ctx, id)
}

func formatUserResponse(u *entity.User) dto.UserResponse {
	return dto.UserResponse{
		ID:        u.ID.Hex(),
		Name:      u.Name,
		Email:     u.Email,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
