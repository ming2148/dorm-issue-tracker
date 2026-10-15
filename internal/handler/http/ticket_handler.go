package http

import (
	"errors"

	"dorm-issue-tracker/internal/domain"
	"dorm-issue-tracker/internal/dto"
	"dorm-issue-tracker/internal/handler/middleware"
	"dorm-issue-tracker/internal/service"
	"dorm-issue-tracker/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TicketHandler struct {
	ticketService service.TicketService
}

func NewTicketHandler(ticketService service.TicketService) *TicketHandler {
	return &TicketHandler{ticketService: ticketService}
}

// Create handles Form 1: Repair Request Form
func (h *TicketHandler) Create(c *gin.Context) {
	var req dto.CreateTicketRequest
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "กรุณากรอกข้อมูลให้ครบถ้วน: "+err.Error())
		return
	}

	form, _ := c.MultipartForm()
	if form != nil {
		req.Images = form.File["images"]
	}

	userID := c.MustGet(middleware.CtxUserIDKey).(uuid.UUID)
	roomNo := c.MustGet(middleware.CtxRoomNoKey).(string)

	ticket, err := h.ticketService.CreateTicket(c.Request.Context(), userID, roomNo, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidFileType) || errors.Is(err, service.ErrFileSizeExceeded) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalServerError(c, "ไม่สามารถสร้างใบแจ้งซ่อมได้: "+err.Error())
		return
	}

	response.Created(c, "ส่งใบแจ้งซ่อมเรียบร้อยแล้ว", ticket)
}

// GetMyTickets lists resident's tickets with status filter
func (h *TicketHandler) GetMyTickets(c *gin.Context) {
	userID := c.MustGet(middleware.CtxUserIDKey).(uuid.UUID)
	status := c.Query("status")

	tickets, err := h.ticketService.GetMyTickets(c.Request.Context(), userID, status)
	if err != nil {
		response.InternalServerError(c, "ไม่สามารถดึงข้อมูลประวัติได้")
		return
	}

	response.Success(c, "ดึงข้อมูลประวัติสำเร็จ", tickets)
}

// GetByID gets a single ticket with IDOR validation
func (h *TicketHandler) GetByID(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "รหัสใบแจ้งซ่อมไม่ถูกต้อง")
		return
	}

	userID := c.MustGet(middleware.CtxUserIDKey).(uuid.UUID)
	role := c.MustGet(middleware.CtxRoleKey).(domain.Role)

	ticket, err := h.ticketService.GetTicketByID(c.Request.Context(), ticketID, userID, role)
	if err != nil {
		if errors.Is(err, service.ErrTicketNotFound) {
			response.NotFound(c, "ไม่พบใบแจ้งซ่อม")
			return
		}
		if errors.Is(err, service.ErrUnauthorizedAccess) {
			response.Forbidden(c, "คุณไม่มีสิทธิ์เข้าถึงใบแจ้งซ่อมนี้")
			return
		}
		response.InternalServerError(c, "เกิดข้อผิดพลาดในการดึงข้อมูล")
		return
	}

	response.Success(c, "ดึงข้อมูลสำเร็จ", ticket)
}

// Update allows resident to update only when status is Pending
func (h *TicketHandler) Update(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "รหัสใบแจ้งซ่อมไม่ถูกต้อง")
		return
	}

	var req dto.UpdateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "ข้อมูลแก้ไขไม่ถูกต้อง: "+err.Error())
		return
	}

	userID := c.MustGet(middleware.CtxUserIDKey).(uuid.UUID)

	ticket, err := h.ticketService.UpdateTicket(c.Request.Context(), ticketID, userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTicketNotFound):
			response.NotFound(c, "ไม่พบใบแจ้งซ่อม")
		case errors.Is(err, service.ErrUnauthorizedAccess):
			response.Forbidden(c, "คุณไม่มีสิทธิ์แก้ไขใบแจ้งซ่อมนี้")
		case errors.Is(err, service.ErrTicketNotPending):
			response.Conflict(c, "ไม่สามารถแก้ไขได้เนื่องจากช่างเริ่มดำเนินงานหรือเสร็จสิ้นแล้ว (แก้ไขได้เฉพาะสถานะรอดำเนินการ)")
		default:
			response.InternalServerError(c, "ไม่สามารถแก้ไขใบแจ้งซ่อมได้")
		}
		return
	}

	response.Success(c, "แก้ไขใบแจ้งซ่อมสำเร็จ", ticket)
}

// Cancel allows resident to cancel only when status is Pending
func (h *TicketHandler) Cancel(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "รหัสใบแจ้งซ่อมไม่ถูกต้อง")
		return
	}

	userID := c.MustGet(middleware.CtxUserIDKey).(uuid.UUID)

	err = h.ticketService.CancelTicket(c.Request.Context(), ticketID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTicketNotFound):
			response.NotFound(c, "ไม่พบใบแจ้งซ่อม")
		case errors.Is(err, service.ErrUnauthorizedAccess):
			response.Forbidden(c, "คุณไม่มีสิทธิ์ยกเลิกใบแจ้งซ่อมนี้")
		case errors.Is(err, service.ErrTicketNotPending):
			response.Conflict(c, "ไม่สามารถยกเลิกได้เนื่องจากงานอยู่ระหว่างดำเนินการหรือเสร็จสิ้นแล้ว")
		default:
			response.InternalServerError(c, "ไม่สามารถยกเลิกใบแจ้งซ่อมได้")
		}
		return
	}

	response.Success(c, "ยกเลิกใบแจ้งซ่อมเรียบร้อยแล้ว", nil)
}
