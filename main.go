package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/lotsoo/anti_bully_backend/config"
	"github.com/lotsoo/anti_bully_backend/handlers"
	"github.com/lotsoo/anti_bully_backend/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, relying on environment variables")
	}

	cfg, err := config.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := config.NewGormDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto-migrate can fail on some environments (driver/version differences).
	// Allow skipping migrations in dev by setting SKIP_MIGRATE=1 in the environment.
	if os.Getenv("SKIP_MIGRATE") == "1" {
		log.Println("SKIP_MIGRATE=1 set, skipping AutoMigrate")
	} else {
		if err := config.AutoMigrate(db); err != nil {
			// Log the error but continue running so other parts of the server can be used while
			// you investigate migration issues. In production you probably want to fail fast.
			log.Printf("auto-migrate failed: %v", err)
		}
	}

	h := handlers.NewHandler(db, cfg)

	// optionally seed an admin user if env vars are provided
	adminUser := os.Getenv("ADMIN_USER")
	adminPass := os.Getenv("ADMIN_PASS")
	if adminUser != "" && adminPass != "" {
		if err := h.Auth.EnsureAdminExists(adminUser, adminPass); err != nil {
			log.Printf("failed to ensure admin user: %v", err)
		}
	}

	r := gin.Default()
	// allow moderately sized multipart form buffers (8 MiB) so we can validate file sizes ourselves
	r.MaxMultipartMemory = 8 << 20 // 8 MiB

	// serve uploaded files
	r.Static("/uploads", cfg.UploadDir)

	// public
	// registration
	r.POST("/auth/register", h.Auth.Register)
	r.POST("/auth/login", h.Auth.Login)
	r.POST("/reports", h.Report.CreateReport)
	r.GET("/reports/:id", h.Report.GetReport)

	// protected
	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware(cfg))
	{
		auth.PUT("/reports/:id/handle", middleware.RequireRole("admin"), h.Report.HandleReport)
		// admin-only list
		auth.GET("/reports", middleware.RequireRole("admin"), h.Report.ListReports)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("starting server on :%s", port)
	r.Run(":" + port)
}
