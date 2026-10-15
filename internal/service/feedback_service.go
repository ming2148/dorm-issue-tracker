package service

import (
	"context"
	"errors"

	"dorm-issue-tracker/internal/domain"
	"dorm-issue-tracker/internal/dto"
	"dorm-issue-tracker/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTicketNotCompleted   = errors.New("cannot submit feedback: ticket status must be 'Completed'")
	ErrFeedbackAlreadyGiven = errors.New("feedback has already been submitted for this ticket")
)

type FeedbackService interface {
	SubmitFeedback(ctx context.Context, ticketID, userID uuid.UUID, req dto.FeedbackRequest) (*domain.Feedback, error)
	GetFeedback(ctx context.Context, ticketID uuid.UUID) (*domain.Feedback, error)
}

type feedbackService struct {
	ticketRepo   repository.TicketRepository
	feedbackRepo repository.FeedbackRepository
	db           *gorm.DB
}

func NewFeedbackService(ticketRepo repository.TicketRepository, feedbackRepo repository.FeedbackRepository, db *gorm.DB) FeedbackService {
	return &feedbackService{
		ticketRepo:   ticketRepo,
		feedbackRepo: feedbackRepo,
		db:           db,
	}
}

func (s *feedbackService) SubmitFeedback(ctx context.Context, ticketID, userID uuid.UUID, req dto.FeedbackRequest) (*domain.Feedback, error) {
	ticket, err := s.ticketRepo.FindByID(ctx, ticketID)
	if err != nil {
		return nil, ErrTicketNotFound
	}

	// 1. ตรวจสอบความเป็นเจ้าของใบแจ้งซ่อม
	if ticket.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}

	// 2. ส่งได้เฉพาะรายการที่สถานะเป็น 'Completed' เท่านั้น
	if ticket.Status != domain.StatusCompleted {
		return nil, ErrTicketNotCompleted
	}

	// 3. ตรวจสอบว่าเคยส่งฟีดแบ็กไปแล้วหรือยัง
	existing, _ := s.feedbackRepo.FindByTicketID(ctx, ticketID)
	if existing != nil {
		return nil, ErrFeedbackAlreadyGiven
	}

	feedback := &domain.Feedback{
		TicketID:       ticketID,
		UserID:         userID,
		Rating:         req.Rating,
		Comment:        req.Comment,
		TechEvaluation: req.TechEvaluation,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(feedback).Error; err != nil {
			return err
		}
		// ปิดจ็อบเป็น Closed อัตโนมัติหลังรับรีวิว
		return tx.Model(ticket).Update("status", domain.StatusClosed).Error
	})
	if err != nil {
		return nil, err
	}

	return feedback, nil
}

func (s *feedbackService) GetFeedback(ctx context.Context, ticketID uuid.UUID) (*domain.Feedback, error) {
	return s.feedbackRepo.FindByTicketID(ctx, ticketID)
}
