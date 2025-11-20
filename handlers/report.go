package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lotsoo/anti_bully_backend/models"
	"gorm.io/gorm"
)

type ReportHandler struct {
	DB  *gorm.DB
	cfg *ConfigWrapper
}

// CreateReport accepts either JSON or multipart/form-data (with field `image` for file).
func (r *ReportHandler) CreateReport(c *gin.Context) {
	var (
		location    string
		description string
		imageURL    string
	)

	ct := c.ContentType()
	if strings.HasPrefix(ct, "multipart/form-data") || strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data") {
		// form upload
		location = c.PostForm("location")
		description = c.PostForm("description")
		file, err := c.FormFile("image")
		if err == nil && file != nil {
			const maxFileBytes = 5 << 20 // 5 MiB

			// quick check if client provided size
			if file.Size > 0 && file.Size > maxFileBytes {
				c.JSON(http.StatusBadRequest, gin.H{"error": "file too large; max 5MB"})
				return
			}

			// validate mime type by reading the first 512 bytes and also ensure size <= max
			fh, err := file.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read uploaded file"})
				return
			}
			defer fh.Close()
			buf := make([]byte, 512)
			n, _ := fh.Read(buf)
			contentType := http.DetectContentType(buf[:n])
			// allow only PNG and JPEG
			if contentType != "image/png" && contentType != "image/jpeg" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image type; only PNG and JPEG allowed"})
				return
			}

			// if file.Size was not provided (0), attempt to read up to max+1 bytes to detect oversize
			if file.Size == 0 {
				// we've already read n bytes; try to read the rest up to (maxFileBytes+1 - n)
				remaining := int64(maxFileBytes) + 1 - int64(n)
				if remaining > 0 {
					// io.CopyN will return nil if it read exactly 'remaining' bytes (meaning file is at least that big)
					_, err := io.CopyN(io.Discard, fh, remaining)
					if err == nil {
						// there was more than maxFileBytes bytes
						c.JSON(http.StatusBadRequest, gin.H{"error": "file too large; max 5MB"})
						return
					}
					// if err is io.EOF, file smaller than limit — ok
				}
			}

			uploadDir := r.cfg.UploadDir
			if uploadDir == "" {
				uploadDir = "./uploads"
			}
			if err := os.MkdirAll(uploadDir, 0o755); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create upload dir"})
				return
			}
			fname := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
			outPath := filepath.Join(uploadDir, fname)
			if err := c.SaveUploadedFile(file, outPath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save uploaded file"})
				return
			}
			imageURL = "/uploads/" + fname
		}
	} else {
		// JSON body
		var req struct {
			Location    string `json:"location" binding:"required"`
			Description string `json:"description" binding:"required"`
			ImageURL    string `json:"image_url"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		location = req.Location
		description = req.Description
		imageURL = req.ImageURL
	}

	if strings.TrimSpace(location) == "" || strings.TrimSpace(description) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "location and description are required"})
		return
	}

	rec := models.Report{
		Location:    location,
		Description: description,
		ImageURL:    imageURL,
		Status:      "BELUM DITANGANI",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// If the request contains authenticated user info (set by middleware), attach reporter info
	if v, ok := c.Get("user_id"); ok {
		if uid, ok2 := v.(uint); ok2 {
			rec.ReporterID = uid
			// try to fetch username
			var u models.User
			if err := r.DB.First(&u, uid).Error; err == nil {
				rec.ReporterName = u.Username
			}
		}
	}
	if err := r.DB.Create(&rec).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create report"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"report": rec})
}

func (r *ReportHandler) GetReport(c *gin.Context) {
	id := c.Param("id")
	var rec models.Report
	if err := r.DB.First(&rec, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"report": rec})
}

func (r *ReportHandler) ListReports(c *gin.Context) {
	var list []models.Report
	if err := r.DB.Order("created_at desc").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"reports": list})
}

// HandleReport marks as handled; admin-only endpoint
func (r *ReportHandler) HandleReport(c *gin.Context) {
	id := c.Param("id")
	var rec models.Report
	if err := r.DB.First(&rec, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	rec.Status = "SUDAH DITANGANI"
	rec.UpdatedAt = time.Now()
	if err := r.DB.Save(&rec).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"report": rec})
}
