package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"dorm-issue-tracker/internal/domain"
	handlerHttp "dorm-issue-tracker/internal/handler/http"
	"dorm-issue-tracker/internal/handler/middleware"
	"dorm-issue-tracker/internal/repository"
	"dorm-issue-tracker/internal/service"
	"dorm-issue-tracker/pkg/database"
	"dorm-issue-tracker/pkg/hasher"
	"dorm-issue-tracker/pkg/jwt"
	"dorm-issue-tracker/web"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func main() {
	log.Println("Starting CondoCare Issue Tracker Application...")

	// 1. Initialize Database
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	// 2. Initialize Core Packages
	pwHasher := hasher.NewBcryptHasher(10)
	jwtSecret := getEnv("JWT_SECRET", "super-secret-key-condocare-2026-production")
	jwtRefreshSecret := getEnv("JWT_REFRESH_SECRET", "refresh-secret-key-condocare-2026")
	jwtService := jwt.NewJWTService(jwtSecret, jwtRefreshSecret, 24*time.Hour, 7*24*time.Hour)

	// 3. Seed Demo Data
	if err := database.SeedData(db, pwHasher); err != nil {
		log.Printf("Warning: failed to seed demo data: %v", err)
	}

	// 4. Initialize Repositories
	userRepo := repository.NewUserRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	feedbackRepo := repository.NewFeedbackRepository(db)

	// 5. Initialize Services
	uploadDir := getEnv("UPLOAD_DIR", "./uploads")
	_ = os.MkdirAll(uploadDir, 0755)

	authService := service.NewAuthService(userRepo, pwHasher, jwtService)
	ticketService := service.NewTicketService(ticketRepo, db, uploadDir)
	feedbackService := service.NewFeedbackService(ticketRepo, feedbackRepo, db)

	// 6. Initialize HTTP Handlers
	authHandler := handlerHttp.NewAuthHandler(authService)
	ticketHandler := handlerHttp.NewTicketHandler(ticketService)
	feedbackHandler := handlerHttp.NewFeedbackHandler(feedbackService)
	adminHandler := handlerHttp.NewAdminHandler(ticketService, userRepo)

	// 7. Setup Router
	r := gin.Default()

	// Global Security Headers
	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Next()
	})

	// Static uploads directory for images
	r.Static("/uploads", uploadDir)

	// Serve Frontend Static Files
	r.GET("/", func(c *gin.Context) {
		indexData, err := web.WebFS.ReadFile("index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load index.html")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexData)
	})

	staticSub, err := fs.Sub(web.WebFS, "static")
	if err != nil {
		log.Fatalf("Failed to create static sub filesystem: %v", err)
	}
	r.StaticFS("/static", http.FS(staticSub))

	// Rate Limiter for Login (5 requests per minute, burst of 10)
	loginLimiter := middleware.NewIPRateLimiter(rate.Every(12*time.Second), 10)

	// API Routes Group
	v1 := r.Group("/api/v1")
	{
		// Public Auth Routes
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", loginLimiter.Limit(), authHandler.Login)
			authGroup.POST("/logout", authHandler.Logout)
			authGroup.GET("/me", middleware.AuthMiddleware(jwtService), authHandler.GetCurrentUser)
		}

		// Authenticated Routes
		authRequired := v1.Group("")
		authRequired.Use(middleware.AuthMiddleware(jwtService))
		{
			// Resident Tickets CRUD (with Ownership enforcement)
			ticketsGroup := authRequired.Group("/tickets")
			{
				ticketsGroup.POST("", middleware.RequireRoles(domain.RoleResident), ticketHandler.Create)
				ticketsGroup.GET("", middleware.RequireRoles(domain.RoleResident), ticketHandler.GetMyTickets)
				ticketsGroup.GET("/:id", ticketHandler.GetByID)
				ticketsGroup.PUT("/:id", middleware.RequireRoles(domain.RoleResident), ticketHandler.Update)
				ticketsGroup.DELETE("/:id", middleware.RequireRoles(domain.RoleResident), ticketHandler.Cancel)

				// Form 2: Repair Feedback (Only for completed status)
				ticketsGroup.POST("/:id/feedback", middleware.RequireRoles(domain.RoleResident), feedbackHandler.Submit)
				ticketsGroup.GET("/:id/feedback", feedbackHandler.GetByTicket)
			}

			// Admin & Technician Routes (RBAC: Admin or Technician)
			adminGroup := authRequired.Group("/admin")
			adminGroup.Use(middleware.RequireRoles(domain.RoleAdmin, domain.RoleTechnician))
			{
				adminGroup.GET("/dashboard", adminHandler.GetDashboard)
				adminGroup.GET("/tickets", adminHandler.ListAllTickets)
				adminGroup.PATCH("/tickets/:id/status", adminHandler.ChangeStatus)
				adminGroup.GET("/technicians", adminHandler.ListTechnicians)
			}
		}
	}

	port := getEnv("PORT", "8080")
	log.Printf("===================================================================")
	log.Printf(" CondoCare Issue Tracker Server listening on http://localhost:%s", port)
	log.Printf(" Demo Accounts available:")
	log.Printf("   1. Resident:   resident@condo.local / password123 (Room 402)")
	log.Printf("   2. Admin/Tech: admin@condo.local    / admin123")
	log.Printf("   3. Technician: tech@condo.local     / tech123")
	log.Printf("===================================================================")

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
