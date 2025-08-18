package handler

import (
	"fmt"
	"net/http"
	"os"

	"go-distributed-hashcat/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WordlistHandler struct {
	wordlistUsecase usecase.WordlistUsecase
}

func NewWordlistHandler(wordlistUsecase usecase.WordlistUsecase) *WordlistHandler {
	return &WordlistHandler{
		wordlistUsecase: wordlistUsecase,
	}
}

func (h *WordlistHandler) UploadWordlist(c *gin.Context) {
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
	wordlist, err := h.wordlistUsecase.UploadWordlist(
		c.Request.Context(),
		file.Filename,
		src,
		file.Size,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": wordlist})
}

func (h *WordlistHandler) GetWordlist(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}

	wordlist, err := h.wordlistUsecase.GetWordlist(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": wordlist})
}

func (h *WordlistHandler) GetAllWordlists(c *gin.Context) {
	wordlists, err := h.wordlistUsecase.GetAllWordlists(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": wordlists})
}

func (h *WordlistHandler) DeleteWordlist(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}

	if err := h.wordlistUsecase.DeleteWordlist(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Wordlist deleted successfully"})
}

func (h *WordlistHandler) GetWordlistContent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}

	wordlist, err := h.wordlistUsecase.GetWordlist(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Read file content
	content, err := os.ReadFile(wordlist.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read wordlist content"})
		return
	}

	// Return content as plain text
	c.Header("Content-Type", "text/plain")
	c.Data(http.StatusOK, "text/plain", content)
}

func (h *WordlistHandler) DownloadWordlist(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}

	wordlist, err := h.wordlistUsecase.GetWordlist(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Set headers for file download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", wordlist.OrigName))
	c.Header("Content-Type", "text/plain")
	c.Header("Content-Transfer-Encoding", "binary")

	// Serve the file
	c.File(wordlist.Path)
}
