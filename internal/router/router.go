package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"monitoring-cctv-be/config"
	"monitoring-cctv-be/internal/handler"
	"monitoring-cctv-be/internal/middleware"
	"monitoring-cctv-be/internal/repository"
	"monitoring-cctv-be/internal/service"
)

func Setup(db *gorm.DB, cfg *config.Config, media service.MediaProvider) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Recovery(), gin.Logger(), middleware.CORS())

	r.GET("/health", handler.Health)

	// Repositories
	userRepo := repository.NewUserRepository(db)
	edgeRepo := repository.NewEdgeRepository(db)
	cameraRepo := repository.NewCameraRepository(db)
	recordingRepo := repository.NewRecordingRepository(db)

	// Services
	accessTTL, _ := time.ParseDuration(cfg.JWTExpiry)
	refreshTTL, _ := time.ParseDuration(cfg.JWTRefreshExpiry)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, accessTTL, refreshTTL)
	userService := service.NewUserService(userRepo)
	edgeService := service.NewEdgeService(edgeRepo)
	cameraService := service.NewCameraService(cameraRepo, edgeRepo, media)
	recordingService := service.NewRecordingService(recordingRepo)
	dashboardService := service.NewDashboardService(db)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	edgeHandler := handler.NewEdgeHandler(edgeService)
	cameraHandler := handler.NewCameraHandler(cameraService)
	recordingHandler := handler.NewRecordingHandler(recordingService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)

	api := r.Group("/api/v1")

	// Public
	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/refresh", authHandler.Refresh)

	// Authenticated (any role)
	auth := api.Group("")
	auth.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		auth.POST("/auth/logout", authHandler.Logout)

		auth.GET("/cameras", cameraHandler.List)
		auth.GET("/edges", edgeHandler.List)
		auth.GET("/edges/:id", edgeHandler.Get)
		auth.GET("/recordings", recordingHandler.List)
		auth.GET("/recordings/cameras", recordingHandler.Cameras)
		auth.GET("/recordings/:id", recordingHandler.Get)
		auth.GET("/admin/summary", dashboardHandler.Summary)
	}

	// Admin + superadmin — edges/cameras/recordings management
	admin := api.Group("")
	admin.Use(middleware.AuthMiddleware(cfg.JWTSecret), middleware.RequireRole("admin", "superadmin"))
	{
		admin.GET("/cameras/:id", cameraHandler.Get) // includes source_url — admin/superadmin only
		admin.POST("/cameras", cameraHandler.Create)
		admin.PATCH("/cameras/:id", cameraHandler.Patch)
		admin.DELETE("/cameras/:id", cameraHandler.Delete)

		admin.POST("/edges", edgeHandler.Create)
		admin.PUT("/edges/:id", edgeHandler.Update)
		admin.DELETE("/edges/:id", edgeHandler.Delete)

		admin.POST("/recordings/archive-bulk", recordingHandler.ArchiveBulk)
	}

	// Superadmin only — user management
	superadmin := api.Group("")
	superadmin.Use(middleware.AuthMiddleware(cfg.JWTSecret), middleware.RequireRole("superadmin"))
	{
		superadmin.GET("/users", userHandler.List)
		superadmin.GET("/users/:id", userHandler.Get)
		superadmin.POST("/users", userHandler.Create)
		superadmin.PUT("/users/:id", userHandler.Update)
		superadmin.PATCH("/users/:id/active", userHandler.SetActive)
		superadmin.DELETE("/users/:id", userHandler.Delete)
		superadmin.POST("/users/:id/reset-password", userHandler.ResetPassword)
	}

	return r
}
