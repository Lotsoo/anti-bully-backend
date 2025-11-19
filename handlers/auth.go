package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lotsoo/anti_bully_backend/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB  *gorm.DB
	cfg *ConfigWrapper // minimal wrapper to access JWT secret if needed
}

// ConfigWrapper keeps a lightweight reference to config.Config without importing the package here
// to avoid circular imports in small scaffolding. We'll define it here and populate from NewHandler.
type ConfigWrapper struct {
	JWTSecret string
	UploadDir string
}

func (a *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := a.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// generate token (we will use utils.GenerateToken via context set in main)
	token, err := GenerateTokenForUser(user.ID, user.Role, a.cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": gin.H{"id": user.ID, "username": user.Username, "role": user.Role}})
}

// Helper to create an admin user if none exist (simple seeding)
func (a *AuthHandler) EnsureAdminExists(username, password string) error {
	var count int64
	a.DB.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	if count > 0 {
		return nil
	}

	pw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u := models.User{Username: username, Password: string(pw), Role: "admin", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return a.DB.Create(&u).Error
}

// GenerateTokenForUser is defined in utils/jwt.go but to avoid import cycle we'll declare here as a wrapper
func GenerateTokenForUser(userID uint, role string, secret string) (string, error) {
	return generateJWT(userID, role, secret)
}
