package middleware

import (
	"strings"

	"dorm-issue-tracker/internal/domain"
	"dorm-issue-tracker/pkg/jwt"
	"dorm-issue-tracker/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	CtxUserIDKey = "current_user_id"
	CtxRoleKey   = "current_user_role"
	CtxRoomNoKey = "current_user_room"
	CtxEmailKey  = "current_user_email"
	CtxNameKey   = "current_user_name"
)

func AuthMiddleware(jwtService jwt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string

		// 1. Try Cookie
		if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
			tokenStr = cookie
		} else {
			// 2. Try Authorization Header
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenStr == "" {
			response.Unauthorized(c, "Missing authentication token")
			c.Abort()
			return
		}

		claims, err := jwtService.ValidateToken(tokenStr)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired session. Please log in again.")
			c.Abort()
			return
		}

		c.Set(CtxUserIDKey, claims.UserID)
		c.Set(CtxRoleKey, domain.Role(claims.Role))
		c.Set(CtxRoomNoKey, claims.RoomNo)
		c.Set(CtxEmailKey, claims.Email)
		c.Set(CtxNameKey, claims.FullName)

		c.Next()
	}
}

func RequireRoles(allowedRoles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get(CtxRoleKey)
		if !exists {
			response.Forbidden(c, "User role not identified")
			c.Abort()
			return
		}

		userRole := roleVal.(domain.Role)
		for _, r := range allowedRoles {
			if userRole == r {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "Access denied: you do not have permission for this resource")
		c.Abort()
	}
}
