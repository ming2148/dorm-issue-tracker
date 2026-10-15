package repository

import (
	"context"
	"dorm-issue-tracker/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FeedbackRepository interface {
	Create(ctx context.Context, feedback *domain.Feedback) error
	FindByTicketID(ctx context.Context, ticketID uuid.UUID) (*domain.Feedback, error)
}

type feedbackRepository struct {
	db *gorm.DB
}

func NewFeedbackRepository(db *gorm.DB) FeedbackRepository {
	return &feedbackRepository{db: db}
}

func (r *feedbackRepository) Create(ctx context.Context, feedback *domain.Feedback) error {
	return r.db.WithContext(ctx).Create(feedback).Error
}

func (r *feedbackRepository) FindByTicketID(ctx context.Context, ticketID uuid.UUID) (*domain.Feedback, error) {
	var feedback domain.Feedback
	err := r.db.WithContext(ctx).Preload("User").First(&feedback, "ticket_id = ?", ticketID).Error
	if err != nil {
		return nil, err
	}
	return &feedback, nil
}
