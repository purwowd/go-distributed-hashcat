package handler

import (
	"fmt"
	"net/http"
	"strings"

	"go-distributed-hashcat/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type HashFileHandler struct {
	hashFileUsecase usecase.HashFileUsecase
}

func NewHashFileHandler(hashFileUsecase usecase.HashFileUsecase) *HashFileHandler {
	return &HashFileHandler{
		hashFileUsecase: hashFileUsecase,
	}
}

func (h *HashFileHandler) UploadHashFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer src.Close()

	// Upload the file
	hashFile, err := h.hashFileUsecase.UploadHashFile(
		c.Request.Context(),
		file.Filename,
		src,
		file.Size,
	)
	if err != nil {
		if strings.Contains(err.Error(), "file already exists") {
			// Extract the filename from the error message
			// Error format: "file already exists: filename.txt"
			parts := strings.Split(err.Error(), ": ")
			if len(parts) == 2 {
				filename := parts[1]
				c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("%s already exists", filename)})
			} else {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload hash file"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": hashFile})
}

func (h *HashFileHandler) GetHashFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hash file ID"})
		return
	}

	hashFile, err := h.hashFileUsecase.GetHashFile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": hashFile})
}

func (h *HashFileHandler) GetAllHashFiles(c *gin.Context) {
	hashFiles, err := h.hashFileUsecase.GetAllHashFiles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": hashFiles})
}

func (h *HashFileHandler) DeleteHashFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hash file ID"})
		return
	}

	if err := h.hashFileUsecase.DeleteHashFile(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Hash file deleted successfully"})
}

func (h *HashFileHandler) DownloadHashFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hash file ID"})
		return
	}

	hashFile, err := h.hashFileUsecase.GetHashFile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Set headers for file download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", hashFile.OrigName))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Transfer-Encoding", "binary")

	// Serve the file
	c.File(hashFile.Path)
}
