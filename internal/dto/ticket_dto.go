package dto

import (
	"dorm-issue-tracker/internal/domain"
	"mime/multipart"
)

type CreateTicketRequest struct {
	Title       string                  `form:"title" binding:"required,min=3,max=150"`
	Category    domain.TicketCategory   `form:"category" binding:"required,oneof=Electrical Plumbing Appliance Structural Other"`
	Description string                  `form:"description" binding:"required,min=5,max=2000"`
	Urgency     domain.UrgencyLevel     `form:"urgency" binding:"required,oneof=Normal Urgent Emergency"`
	Images      []*multipart.FileHeader `form:"images"`
}

type UpdateTicketRequest struct {
	Title       string                `json:"title" binding:"required,min=3,max=150"`
	Category    domain.TicketCategory `json:"category" binding:"required,oneof=Electrical Plumbing Appliance Structural Other"`
	Description string                `json:"description" binding:"required,min=5,max=2000"`
	Urgency     domain.UrgencyLevel   `json:"urgency" binding:"required,oneof=Normal Urgent Emergency"`
}

type ChangeStatusRequest struct {
	Status       domain.TicketStatus `json:"status" binding:"required,oneof=Pending In_Progress Completed Closed Cancelled"`
	TechnicianID *string             `json:"technician_id" binding:"omitempty"`
	InternalNote string              `json:"internal_note"`
}

type FeedbackRequest struct {
	Rating         int    `json:"rating" binding:"required,min=1,max=5"`
	Comment        string `json:"comment"`
	TechEvaluation string `json:"tech_evaluation"`
}

type DashboardStats struct {
	TotalTickets      int64            `json:"total_tickets"`
	PendingCount      int64            `json:"pending_count"`
	InProgressCount   int64            `json:"in_progress_count"`
	CompletedCount    int64            `json:"completed_count"`
	CancelledCount    int64            `json:"cancelled_count"`
	CategoryBreakdown map[string]int64 `json:"category_breakdown"`
}
