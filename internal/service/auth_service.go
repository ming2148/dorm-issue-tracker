package service

import (
	"context"
	"errors"
	"strings"

	"dorm-issue-tracker/internal/domain"
	"dorm-issue-tracker/internal/dto"
	"dorm-issue-tracker/internal/repository"
	"dorm-issue-tracker/pkg/hasher"
	"dorm-issue-tracker/pkg/jwt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email/room number or password")
	ErrUserNotFound       = errors.New("user not found")
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, string, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*dto.UserProfile, error)
}

type authService struct {
	userRepo   repository.UserRepository
	hasher     hasher.PasswordHasher
	jwtService jwt.Service
}

func NewAuthService(userRepo repository.UserRepository, hasher hasher.PasswordHasher, jwtService jwt.Service) AuthService {
	return &authService{
		userRepo:   userRepo,
		hasher:     hasher,
		jwtService: jwtService,
	}
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	existing, _ := s.userRepo.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	hash, err := s.hasher.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		FullName:     strings.TrimSpace(req.FullName),
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		PhoneNumber:  strings.TrimSpace(req.PhoneNumber),
		RoomNo:       strings.TrimSpace(req.RoomNo),
		PasswordHash: hash,
		Role:         domain.RoleResident,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	accessToken, _, err := s.jwtService.GenerateTokens(user.ID, user.Email, user.FullName, string(user.Role), user.RoomNo)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken: accessToken,
		User: dto.UserProfile{
			ID:          user.ID.String(),
			FullName:    user.FullName,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			RoomNo:      user.RoomNo,
			Role:        string(user.Role),
		},
	}, nil
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, string, error) {
	identifier := strings.TrimSpace(req.Identifier)
	user, err := s.userRepo.FindByEmailOrRoom(ctx, strings.ToLower(identifier))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if !s.hasher.CheckPassword(req.Password, user.PasswordHash) {
		return nil, "", ErrInvalidCredentials
	}

	accessToken, refreshToken, err := s.jwtService.GenerateTokens(user.ID, user.Email, user.FullName, string(user.Role), user.RoomNo)
	if err != nil {
		return nil, "", err
	}

	return &dto.AuthResponse{
		AccessToken: accessToken,
		User: dto.UserProfile{
			ID:          user.ID.String(),
			FullName:    user.FullName,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			RoomNo:      user.RoomNo,
			Role:        string(user.Role),
		},
	}, refreshToken, nil
}

func (s *authService) GetProfile(ctx context.Context, userID uuid.UUID) (*dto.UserProfile, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return &dto.UserProfile{
		ID:          user.ID.String(),
		FullName:    user.FullName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		RoomNo:      user.RoomNo,
		Role:        string(user.Role),
	}, nil
}
