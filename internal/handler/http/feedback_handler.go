package http

import (
	"errors"

	"dorm-issue-tracker/internal/dto"
	"dorm-issue-tracker/internal/handler/middleware"
	"dorm-issue-tracker/internal/service"
	"dorm-issue-tracker/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FeedbackHandler struct {
	feedbackService service.FeedbackService
}

func NewFeedbackHandler(feedbackService service.FeedbackService) *FeedbackHandler {
	return &FeedbackHandler{feedbackService: feedbackService}
}

// Submit handles Form 2: Repair Feedback Form
func (h *FeedbackHandler) Submit(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "รหัสใบแจ้งซ่อมไม่ถูกต้อง")
		return
	}

	var req dto.FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "กรุณาให้คะแนน 1-5 ดาว: "+err.Error())
		return
	}

	userID := c.MustGet(middleware.CtxUserIDKey).(uuid.UUID)

	feedback, err := h.feedbackService.SubmitFeedback(c.Request.Context(), ticketID, userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTicketNotFound):
			response.NotFound(c, "ไม่พบใบแจ้งซ่อม")
		case errors.Is(err, service.ErrUnauthorizedAccess):
			response.Forbidden(c, "คุณไม่มีสิทธิ์ส่งฟีดแบ็กสำหรับใบแจ้งซ่อมนี้")
		case errors.Is(err, service.ErrTicketNotCompleted):
			response.Conflict(c, "สามารถส่งรีวิวได้เฉพาะงานซ่อมที่สถานะเป็น 'Completed (เสร็จสิ้น)' เท่านั้น")
		case errors.Is(err, service.ErrFeedbackAlreadyGiven):
			response.Conflict(c, "คุณได้ส่งรีวิวสำหรับรายการนี้ไปแล้ว")
		default:
			response.InternalServerError(c, "ไม่สามารถส่งแบบประเมินได้: "+err.Error())
		}
		return
	}

	response.Created(c, "ขอบคุณสำหรับการประเมินความพึงพอใจ", feedback)
}

func (h *FeedbackHandler) GetByTicket(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "รหัสใบแจ้งซ่อมไม่ถูกต้อง")
		return
	}

	feedback, err := h.feedbackService.GetFeedback(c.Request.Context(), ticketID)
	if err != nil {
		response.NotFound(c, "ยังไม่มีข้อมูลการประเมิน")
		return
	}

	response.Success(c, "ข้อมูลการประเมิน", feedback)
}
