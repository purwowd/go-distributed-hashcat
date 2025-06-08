package usecase

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go-distributed-hashcat/internal/domain"

	"github.com/google/uuid"
)

type HashFileUsecase interface {
	UploadHashFile(ctx context.Context, name string, content io.Reader, size int64) (*domain.HashFile, error)
	GetHashFile(ctx context.Context, id uuid.UUID) (*domain.HashFile, error)
	GetAllHashFiles(ctx context.Context) ([]domain.HashFile, error)
	DeleteHashFile(ctx context.Context, id uuid.UUID) error
}

type hashFileUsecase struct {
	hashFileRepo domain.HashFileRepository
	uploadDir    string
}

func NewHashFileUsecase(hashFileRepo domain.HashFileRepository, uploadDir string) HashFileUsecase {
	return &hashFileUsecase{
		hashFileRepo: hashFileRepo,
		uploadDir:    uploadDir,
	}
}

func (u *hashFileUsecase) UploadHashFile(ctx context.Context, name string, content io.Reader, size int64) (*domain.HashFile, error) {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(u.uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	fileID := uuid.New()
	ext := filepath.Ext(name)
	filename := fmt.Sprintf("%s%s", fileID.String(), ext)
	filePath := filepath.Join(u.uploadDir, filename)

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content to file
	written, err := io.Copy(file, content)
	if err != nil {
		// Clean up on error
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Determine file type
	fileType := u.determineFileType(name)

	// Create hash file record
	hashFile := &domain.HashFile{
		ID:       fileID,
		Name:     filename,
		OrigName: name,
		Path:     filePath,
		Size:     written,
		Type:     fileType,
	}

	if err := u.hashFileRepo.Create(ctx, hashFile); err != nil {
		// Clean up on error
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to create hash file record: %w", err)
	}

	return hashFile, nil
}

func (u *hashFileUsecase) GetHashFile(ctx context.Context, id uuid.UUID) (*domain.HashFile, error) {
	hashFile, err := u.hashFileRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get hash file: %w", err)
	}
	return hashFile, nil
}

func (u *hashFileUsecase) GetAllHashFiles(ctx context.Context) ([]domain.HashFile, error) {
	hashFiles, err := u.hashFileRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get hash files: %w", err)
	}
	return hashFiles, nil
}

func (u *hashFileUsecase) DeleteHashFile(ctx context.Context, id uuid.UUID) error {
	hashFile, err := u.hashFileRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get hash file: %w", err)
	}

	// Delete the physical file
	if err := os.Remove(hashFile.Path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete physical file: %w", err)
	}

	// Delete the record
	if err := u.hashFileRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete hash file record: %w", err)
	}

	return nil
}

func (u *hashFileUsecase) determineFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".hccapx":
		return "hccapx"
	case ".hccap":
		return "hccap"
	case ".cap":
		return "cap"
	case ".pcap":
		return "pcap"
	default:
		return "hash"
	}
}
