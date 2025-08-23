package handler

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"go-distributed-hashcat/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WordlistHandler struct {
	wordlistUsecase usecase.WordlistUsecase
	// Add upload session management for chunked uploads
	uploadSessions map[string]*UploadSession
	sessionMutex   sync.RWMutex
}

// UploadSession represents a chunked upload session
type UploadSession struct {
	ID           string
	Filename     string
	TotalSize    int64
	TotalChunks  int
	ChunkSize    int
	Chunks       map[int][]byte
	UploadedSize int64
	CreatedAt    int64
}

func NewWordlistHandler(wordlistUsecase usecase.WordlistUsecase) *WordlistHandler {
	return &WordlistHandler{
		wordlistUsecase: wordlistUsecase,
		uploadSessions:  make(map[string]*UploadSession),
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

// New: Initialize chunked upload session
func (h *WordlistHandler) InitChunkedUpload(c *gin.Context) {
	var req struct {
		Filename    string `json:"filename" binding:"required"`
		TotalSize   int64  `json:"totalSize" binding:"required"`
		TotalChunks int    `json:"totalChunks" binding:"required"`
		ChunkSize   int    `json:"chunkSize" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Validate file size (allow up to 10GB)
	if req.TotalSize > 10*1024*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 10GB limit"})
		return
	}

	// Create upload session
	uploadID := uuid.New().String()
	session := &UploadSession{
		ID:          uploadID,
		Filename:    req.Filename,
		TotalSize:   req.TotalSize,
		TotalChunks: req.TotalChunks,
		ChunkSize:   req.ChunkSize,
		Chunks:      make(map[int][]byte),
		CreatedAt:   time.Now().Unix(),
	}

	h.sessionMutex.Lock()
	h.uploadSessions[uploadID] = session
	h.sessionMutex.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"uploadId": uploadID,
		"message":  "Upload session initialized",
	})
}

// New: Upload chunk
func (h *WordlistHandler) UploadChunk(c *gin.Context) {
	uploadID := c.PostForm("uploadId")
	chunkIndexStr := c.PostForm("chunkIndex")
	totalChunksStr := c.PostForm("totalChunks")

	if uploadID == "" || chunkIndexStr == "" || totalChunksStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameters"})
		return
	}

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chunk index"})
		return
	}

	totalChunks, err := strconv.Atoi(totalChunksStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid total chunks"})
		return
	}

	// Get upload session
	h.sessionMutex.RLock()
	session, exists := h.uploadSessions[uploadID]
	h.sessionMutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Upload session not found"})
		return
	}

	// Validate chunk index
	if chunkIndex < 0 || chunkIndex >= totalChunks {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chunk index"})
		return
	}

	// Get chunk file
	chunkFile, err := c.FormFile("chunk")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No chunk file uploaded"})
		return
	}

	// Read chunk data
	src, err := chunkFile.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open chunk file"})
		return
	}
	defer src.Close()

	chunkData := make([]byte, chunkFile.Size)
	_, err = src.Read(chunkData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read chunk data"})
		return
	}

	// Store chunk
	h.sessionMutex.Lock()
	session.Chunks[chunkIndex] = chunkData
	session.UploadedSize += int64(len(chunkData))
	h.sessionMutex.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Chunk %d/%d uploaded successfully", chunkIndex+1, totalChunks),
		"progress": float64(len(session.Chunks)) / float64(session.TotalChunks) * 100,
	})
}

// New: Finalize chunked upload
func (h *WordlistHandler) FinalizeChunkedUpload(c *gin.Context) {
	var req struct {
		UploadID string `json:"uploadId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Get upload session
	h.sessionMutex.RLock()
	session, exists := h.uploadSessions[req.UploadID]
	h.sessionMutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Upload session not found"})
		return
	}

	// Check if all chunks are uploaded
	if len(session.Chunks) != session.TotalChunks {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Not all chunks uploaded. Got %d/%d", len(session.Chunks), session.TotalChunks)})
		return
	}

	// Combine chunks and create final file
	combinedData := make([]byte, 0, session.TotalSize)
	for i := 0; i < session.TotalChunks; i++ {
		chunk, exists := session.Chunks[i]
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Missing chunk %d", i)})
			return
		}
		combinedData = append(combinedData, chunk...)
	}

	// Validate total size
	if int64(len(combinedData)) != session.TotalSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Combined file size doesn't match expected size"})
		return
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", "chunked_upload_*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temporary file"})
		return
	}
	defer os.Remove(tempFile.Name())

	// Write combined data
	_, err = tempFile.Write(combinedData)
	if err != nil {
		tempFile.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write combined file"})
		return
	}
	tempFile.Close()

	// Open temp file for reading
	tempFile, err = os.Open(tempFile.Name())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open combined file"})
		return
	}
	defer tempFile.Close()

	// Upload using existing usecase
	wordlist, err := h.wordlistUsecase.UploadWordlist(
		c.Request.Context(),
		session.Filename,
		tempFile,
		session.TotalSize,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to process uploaded file: %v", err)})
		return
	}

	// Clean up session
	h.sessionMutex.Lock()
	delete(h.uploadSessions, req.UploadID)
	h.sessionMutex.Unlock()

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

	err = h.wordlistUsecase.DeleteWordlist(c.Request.Context(), id)
	if err != nil {
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
