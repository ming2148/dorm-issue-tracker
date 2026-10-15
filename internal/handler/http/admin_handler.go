package http

import (
	"errors"

	"dorm-issue-tracker/internal/domain"
	"dorm-issue-tracker/internal/dto"
	"dorm-issue-tracker/internal/repository"
	"dorm-issue-tracker/internal/service"
	"dorm-issue-tracker/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	ticketService service.TicketService
	userRepo      repository.UserRepository
}

func NewAdminHandler(ticketService service.TicketService, userRepo repository.UserRepository) *AdminHandler {
	return &AdminHandler{
		ticketService: ticketService,
		userRepo:      userRepo,
	}
}

// GetDashboard returns overview cards and category breakdowns
func (h *AdminHandler) GetDashboard(c *gin.Context) {
	stats, err := h.ticketService.GetDashboardStats(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "ไม่สามารถโหลดข้อมูลสถิติได้")
		return
	}

	response.Success(c, "ข้อมูลแดชบอร์ด", stats)
}

// ListAllTickets handles search, filter by room/category/status, and sorting
func (h *AdminHandler) ListAllTickets(c *gin.Context) {
	filter := repository.TicketFilter{
		Status:   c.Query("status"),
		Category: c.Query("category"),
		RoomNo:   c.Query("room_no"),
		Search:   c.Query("search"),
	}

	tickets, err := h.ticketService.ListAllTickets(c.Request.Context(), filter)
	if err != nil {
		response.InternalServerError(c, "ไม่สามารถโหลดรายการแจ้งซ่อมได้")
		return
	}

	response.Success(c, "รายการแจ้งซ่อมทั้งหมด", tickets)
}

// ChangeStatus handles action modal: Change status, assign technician, and internal note
func (h *AdminHandler) ChangeStatus(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "รหัสใบแจ้งซ่อมไม่ถูกต้อง")
		return
	}

	var req dto.ChangeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "ข้อมูลสถานะไม่ถูกต้อง: "+err.Error())
		return
	}

	ticket, err := h.ticketService.ChangeStatus(c.Request.Context(), ticketID, req)
	if err != nil {
		if errors.Is(err, service.ErrTicketNotFound) {
			response.NotFound(c, "ไม่พบใบแจ้งซ่อม")
			return
		}
		if errors.Is(err, service.ErrInvalidStatusFlow) {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalServerError(c, "ไม่สามารถเปลี่ยนสถานะได้: "+err.Error())
		return
	}

	response.Success(c, "อัปเดตสถานะและข้อมูลการซ่อมเรียบร้อย", ticket)
}

// ListTechnicians returns list of technicians for assignment dropdown
func (h *AdminHandler) ListTechnicians(c *gin.Context) {
	techs, err := h.userRepo.ListTechnicians(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "ไม่สามารถโหลดรายชื่อช่างได้")
		return
	}

	type TechResponse struct {
		ID       string      `json:"id"`
		FullName string      `json:"full_name"`
		Role     domain.Role `json:"role"`
	}

	var result []TechResponse
	for _, t := range techs {
		result = append(result, TechResponse{
			ID:       t.ID.String(),
			FullName: t.FullName,
			Role:     t.Role,
		})
	}

	response.Success(c, "รายชื่อช่าง", result)
}
