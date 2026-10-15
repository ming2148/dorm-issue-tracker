package repository

import (
	"context"
	"dorm-issue-tracker/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByEmailOrRoom(ctx context.Context, identifier string) (*domain.User, error)
	ListTechnicians(ctx context.Context) ([]domain.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmailOrRoom(ctx context.Context, identifier string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("email = ? OR room_no = ?", identifier, identifier).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) ListTechnicians(ctx context.Context) ([]domain.User, error) {
	var techs []domain.User
	err := r.db.WithContext(ctx).Where("role IN ?", []string{string(domain.RoleTechnician), string(domain.RoleAdmin)}).Find(&techs).Error
	return techs, err
}
