package repository

import (
	"context"
	"dorm-issue-tracker/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketFilter struct {
	Status   string
	Category string
	RoomNo   string
	Search   string
}

type TicketRepository interface {
	Create(ctx context.Context, ticket *domain.Ticket) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Ticket, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, status string) ([]domain.Ticket, error)
	ListAll(ctx context.Context, filter TicketFilter) ([]domain.Ticket, error)
	Update(ctx context.Context, ticket *domain.Ticket) error
	GetStatusCounts(ctx context.Context) (map[string]int64, error)
	GetCategoryCounts(ctx context.Context) (map[string]int64, error)
}

type ticketRepository struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) TicketRepository {
	return &ticketRepository{db: db}
}

func (r *ticketRepository) Create(ctx context.Context, ticket *domain.Ticket) error {
	return r.db.WithContext(ctx).Create(ticket).Error
}

func (r *ticketRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	var ticket domain.Ticket
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("AssignedTechnician").
		Preload("Images").
		Preload("Feedback").
		First(&ticket, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepository) FindByUserID(ctx context.Context, userID uuid.UUID, status string) ([]domain.Ticket, error) {
	var tickets []domain.Ticket
	query := r.db.WithContext(ctx).
		Preload("Images").
		Preload("Feedback").
		Preload("AssignedTechnician").
		Where("user_id = ?", userID).
		Order("created_at DESC")

	if status != "" && status != "All" {
		query = query.Where("status = ?", status)
	}

	err := query.Find(&tickets).Error
	return tickets, err
}

func (r *ticketRepository) ListAll(ctx context.Context, filter TicketFilter) ([]domain.Ticket, error) {
	var tickets []domain.Ticket
	query := r.db.WithContext(ctx).
		Preload("User").
		Preload("AssignedTechnician").
		Preload("Images").
		Preload("Feedback").
		Order("created_at DESC")

	if filter.Status != "" && filter.Status != "All" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Category != "" && filter.Category != "All" {
		query = query.Where("category = ?", filter.Category)
	}
	if filter.RoomNo != "" {
		query = query.Where("room_no LIKE ?", "%"+filter.RoomNo+"%")
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		query = query.Where("title LIKE ? OR description LIKE ? OR ticket_no LIKE ? OR room_no LIKE ?", s, s, s, s)
	}

	err := query.Find(&tickets).Error
	return tickets, err
}

func (r *ticketRepository) Update(ctx context.Context, ticket *domain.Ticket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}

func (r *ticketRepository) GetStatusCounts(ctx context.Context) (map[string]int64, error) {
	type Result struct {
		Status string
		Count  int64
	}
	var results []Result
	err := r.db.WithContext(ctx).Model(&domain.Ticket{}).Select("status, count(*) as count").Group("status").Scan(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, res := range results {
		counts[res.Status] = res.Count
	}
	return counts, nil
}

func (r *ticketRepository) GetCategoryCounts(ctx context.Context) (map[string]int64, error) {
	type Result struct {
		Category string
		Count    int64
	}
	var results []Result
	err := r.db.WithContext(ctx).Model(&domain.Ticket{}).Select("category, count(*) as count").Group("category").Scan(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, res := range results {
		counts[res.Category] = res.Count
	}
	return counts, nil
}
