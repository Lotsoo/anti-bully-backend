package handlers

import (
	"github.com/lotsoo/anti_bully_backend/config"
	"gorm.io/gorm"
)

type Handler struct {
	DB     *gorm.DB
	Config *config.Config
	Auth   *AuthHandler
	Report *ReportHandler
}

func NewHandler(db *gorm.DB, cfg *config.Config) *Handler {
	h := &Handler{DB: db, Config: cfg}
	// create a lightweight config wrapper for handlers to avoid circular imports
	cw := &ConfigWrapper{JWTSecret: cfg.JWTSecret, UploadDir: cfg.UploadDir}
	h.Auth = &AuthHandler{DB: db, cfg: cw}
	h.Report = &ReportHandler{DB: db, cfg: cw}
	return h
}
