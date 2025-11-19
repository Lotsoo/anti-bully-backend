package config

import (
	"errors"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/lotsoo/anti_bully_backend/models"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	UploadDir   string
}

func LoadConfigFromEnv() (*Config, error) {
	db := os.Getenv("DATABASE_URL")
	jwt := os.Getenv("JWT_SECRET")
	upload := os.Getenv("UPLOAD_DIR")
	if upload == "" {
		upload = "./uploads"
	}
	if db == "" || jwt == "" {
		return nil, errors.New("DATABASE_URL and JWT_SECRET must be set")
	}
	return &Config{
		DatabaseURL: db,
		JWTSecret:   jwt,
		UploadDir:   upload,
	}, nil
}

func NewGormDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{}, &models.Report{})
}

func DSNFromEnv() string {
	return os.Getenv("DATABASE_URL")
}

func MustGetJWTSecret() string {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		panic("JWT_SECRET not set")
	}
	return s
}
