package http

import (
	"errors"
	"net/http"

	"dorm-issue-tracker/internal/dto"
	"dorm-issue-tracker/internal/handler/middleware"
	"dorm-issue-tracker/internal/service"
	"dorm-issue-tracker/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "ข้อมูลไม่ถูกต้อง: "+err.Error())
		return
	}

	authResp, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			response.Conflict(c, "อีเมลนี้ถูกใช้งานแล้วในระบบ")
			return
		}
		response.InternalServerError(c, "ไม่สามารถลงทะเบียนได้: "+err.Error())
		return
	}

	// Set HttpOnly Cookie
	c.SetCookie("access_token", authResp.AccessToken, 86400, "/", "", false, true)

	response.Created(c, "ลงทะเบียนสำเร็จเรียบร้อย", authResp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "กรุณากรอกอีเมล/เลขห้อง และรหัสผ่าน")
		return
	}

	authResp, refreshToken, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Unauthorized(c, "อีเมล/หมายเลขห้อง หรือรหัสผ่านไม่ถูกต้อง")
			return
		}
		response.InternalServerError(c, "เกิดข้อผิดพลาดในการเข้าสู่ระบบ")
		return
	}

	// Set HttpOnly Secure Cookies
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", authResp.AccessToken, 86400*1, "/", "", false, true)
	c.SetCookie("refresh_token", refreshToken, 86400*7, "/", "", false, true)

	response.Success(c, "เข้าสู่ระบบสำเร็จ", authResp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	response.Success(c, "ออกจากระบบเรียบร้อย", nil)
}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID := c.MustGet(middleware.CtxUserIDKey).(uuid.UUID)
	profile, err := h.authService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "ไม่พบข้อมูลผู้ใช้งาน")
		return
	}

	response.Success(c, "ข้อมูลผู้ใช้งาน", profile)
}
