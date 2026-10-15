package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"dorm-issue-tracker/internal/domain"
	"dorm-issue-tracker/internal/dto"
	"dorm-issue-tracker/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTicketNotFound      = errors.New("ticket not found")
	ErrUnauthorizedAccess  = errors.New("forbidden: you do not own this ticket")
	ErrTicketNotPending    = errors.New("action not allowed: ticket is already in progress or completed")
	ErrInvalidStatusFlow   = errors.New("invalid status transition")
	ErrInvalidFileType     = errors.New("invalid file type: only jpg, jpeg, and png are allowed")
	ErrFileSizeExceeded    = errors.New("file size exceeds maximum limit of 5MB")
)

type TicketService interface {
	CreateTicket(ctx context.Context, userID uuid.UUID, roomNo string, req dto.CreateTicketRequest) (*domain.Ticket, error)
	GetMyTickets(ctx context.Context, userID uuid.UUID, status string) ([]domain.Ticket, error)
	GetTicketByID(ctx context.Context, ticketID uuid.UUID, requestingUserID uuid.UUID, requestingRole domain.Role) (*domain.Ticket, error)
	ListAllTickets(ctx context.Context, filter repository.TicketFilter) ([]domain.Ticket, error)
	UpdateTicket(ctx context.Context, ticketID, userID uuid.UUID, req dto.UpdateTicketRequest) (*domain.Ticket, error)
	CancelTicket(ctx context.Context, ticketID, userID uuid.UUID) error
	ChangeStatus(ctx context.Context, ticketID uuid.UUID, req dto.ChangeStatusRequest) (*domain.Ticket, error)
	GetDashboardStats(ctx context.Context) (*dto.DashboardStats, error)
}

type ticketService struct {
	ticketRepo repository.TicketRepository
	db         *gorm.DB
	uploadPath string
}

func NewTicketService(ticketRepo repository.TicketRepository, db *gorm.DB, uploadPath string) TicketService {
	_ = os.MkdirAll(uploadPath, 0755)
	return &ticketService{
		ticketRepo: ticketRepo,
		db:         db,
		uploadPath: uploadPath,
	}
}

func (s *ticketService) CreateTicket(ctx context.Context, userID uuid.UUID, roomNo string, req dto.CreateTicketRequest) (*domain.Ticket, error) {
	var images []domain.TicketImage

	for _, fileHeader := range req.Images {
		if fileHeader == nil {
			continue
		}
		if fileHeader.Size > 5*1024*1024 {
			return nil, ErrFileSizeExceeded
		}

		img, err := s.saveUploadedFile(fileHeader)
		if err != nil {
			return nil, err
		}
		images = append(images, *img)
	}

	ticket := &domain.Ticket{
		UserID:      userID,
		RoomNo:      roomNo,
		Title:       strings.TrimSpace(req.Title),
		Category:    req.Category,
		Description: strings.TrimSpace(req.Description),
		Urgency:     req.Urgency,
		Status:      domain.StatusPending,
		Images:      images,
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Create(ticket).Error
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	return ticket, nil
}

func (s *ticketService) UpdateTicket(ctx context.Context, ticketID, userID uuid.UUID, req dto.UpdateTicketRequest) (*domain.Ticket, error) {
	ticket, err := s.ticketRepo.FindByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}

	// 1. IDOR Prevention: ตรวจสอบความเป็นเจ้าของ
	if ticket.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}

	// 2. Business Rule: แก้ไขได้เฉพาะเมื่อสถานะเป็น "Pending" เท่านั้น
	if ticket.Status != domain.StatusPending {
		return nil, ErrTicketNotPending
	}

	ticket.Title = strings.TrimSpace(req.Title)
	ticket.Category = req.Category
	ticket.Description = strings.TrimSpace(req.Description)
	ticket.Urgency = req.Urgency

	if err := s.ticketRepo.Update(ctx, ticket); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (s *ticketService) CancelTicket(ctx context.Context, ticketID, userID uuid.UUID) error {
	ticket, err := s.ticketRepo.FindByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTicketNotFound
		}
		return err
	}

	if ticket.UserID != userID {
		return ErrUnauthorizedAccess
	}

	if ticket.Status != domain.StatusPending {
		return ErrTicketNotPending
	}

	ticket.Status = domain.StatusCancelled
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(ticket).Error; err != nil {
			return err
		}
		return tx.Delete(ticket).Error
	})
}

func (s *ticketService) ChangeStatus(ctx context.Context, ticketID uuid.UUID, req dto.ChangeStatusRequest) (*domain.Ticket, error) {
	ticket, err := s.ticketRepo.FindByID(ctx, ticketID)
	if err != nil {
		return nil, ErrTicketNotFound
	}

	if !isValidTransition(ticket.Status, req.Status) {
		return nil, fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidStatusFlow, ticket.Status, req.Status)
	}

	ticket.Status = req.Status
	if req.InternalNote != "" {
		ticket.InternalNote = req.InternalNote
	}
	if req.TechnicianID != nil && *req.TechnicianID != "" {
		parsedTechID, err := uuid.Parse(*req.TechnicianID)
		if err == nil {
			ticket.AssignedTechnicianID = &parsedTechID
		}
	}

	if err := s.ticketRepo.Update(ctx, ticket); err != nil {
		return nil, err
	}

	return ticket, nil
}

func isValidTransition(from, to domain.TicketStatus) bool {
	transitions := map[domain.TicketStatus][]domain.TicketStatus{
		domain.StatusPending:    {domain.StatusInProgress, domain.StatusCancelled},
		domain.StatusInProgress: {domain.StatusCompleted, domain.StatusPending, domain.StatusCancelled},
		domain.StatusCompleted:  {domain.StatusClosed, domain.StatusInProgress},
		domain.StatusClosed:     {domain.StatusInProgress},
		domain.StatusCancelled:  {domain.StatusPending},
	}

	allowed, ok := transitions[from]
	if !ok {
		return false
	}
	for _, target := range allowed {
		if target == to {
			return true
		}
	}
	return false
}

func (s *ticketService) saveUploadedFile(fileHeader *multipart.FileHeader) (*domain.TicketImage, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// ตรวจสอบ Magic Bytes แรกเพื่อความปลอดภัย (Prevent MIME Spoofing)
	buffer := make([]byte, 512)
	n, err := src.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}

	mimeType := http.DetectContentType(buffer[:n])
	allowedMimes := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
	}

	ext, allowed := allowedMimes[mimeType]
	if !allowed {
		return nil, ErrInvalidFileType
	}

	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	newFilename := uuid.New().String() + ext
	savePath := filepath.Join(s.uploadPath, newFilename)

	dst, err := os.Create(savePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return nil, err
	}

	return &domain.TicketImage{
		FileURL:  "/uploads/" + newFilename,
		FileKey:  savePath,
		MimeType: mimeType,
		FileSize: fileHeader.Size,
	}, nil
}

func (s *ticketService) GetMyTickets(ctx context.Context, userID uuid.UUID, status string) ([]domain.Ticket, error) {
	return s.ticketRepo.FindByUserID(ctx, userID, status)
}

func (s *ticketService) GetTicketByID(ctx context.Context, ticketID uuid.UUID, requestingUserID uuid.UUID, requestingRole domain.Role) (*domain.Ticket, error) {
	ticket, err := s.ticketRepo.FindByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if requestingRole == domain.RoleResident && ticket.UserID != requestingUserID {
		return nil, ErrUnauthorizedAccess
	}

	return ticket, nil
}

func (s *ticketService) ListAllTickets(ctx context.Context, filter repository.TicketFilter) ([]domain.Ticket, error) {
	return s.ticketRepo.ListAll(ctx, filter)
}

func (s *ticketService) GetDashboardStats(ctx context.Context) (*dto.DashboardStats, error) {
	statusCounts, err := s.ticketRepo.GetStatusCounts(ctx)
	if err != nil {
		return nil, err
	}

	categoryCounts, err := s.ticketRepo.GetCategoryCounts(ctx)
	if err != nil {
		return nil, err
	}

	var total int64
	for _, c := range statusCounts {
		total += c
	}

	return &dto.DashboardStats{
		TotalTickets:      total,
		PendingCount:      statusCounts[string(domain.StatusPending)],
		InProgressCount:   statusCounts[string(domain.StatusInProgress)],
		CompletedCount:    statusCounts[string(domain.StatusCompleted)],
		CancelledCount:    statusCounts[string(domain.StatusCancelled)],
		CategoryBreakdown: categoryCounts,
	}, nil
}
