package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Email     string `gorm:"uniqueIndex;size:100" json:"email"`
	Password  string `gorm:"size:255" json:"-"`
	Role      string `gorm:"size:50" json:"role"` // e.g., admin, user
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Report struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Location     string `gorm:"size:255" json:"location"`
	Description  string `gorm:"type:text" json:"description"`
	ImageURL     string `gorm:"size:1024" json:"image_url"`
	ReporterID    uint   `json:"reporter_id"`
	ReporterEmail string `gorm:"size:100" json:"reporter_email"`
	Status        string `gorm:"size:50" json:"status"` // BELUM DITANGANI / SUDAH DITANGANI
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Ensure GORM relationships or helper functions if needed
func (r *Report) BeforeCreate(tx *gorm.DB) (err error) {
	if r.Status == "" {
		r.Status = "BELUM DITANGANI"
	}
	return nil
}
