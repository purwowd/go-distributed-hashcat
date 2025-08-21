package usecase

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go-distributed-hashcat/internal/domain"

	"github.com/google/uuid"
)

type WordlistUsecase interface {
	UploadWordlist(ctx context.Context, name string, content io.Reader, size int64) (*domain.Wordlist, error)
	GetWordlist(ctx context.Context, id uuid.UUID) (*domain.Wordlist, error)
	GetByOrigName(ctx context.Context, origName string) (*domain.Wordlist, error)
	GetAllWordlists(ctx context.Context) ([]domain.Wordlist, error)
	DeleteWordlist(ctx context.Context, id uuid.UUID) error
}

type wordlistUsecase struct {
	wordlistRepo domain.WordlistRepository
	uploadDir    string
}

func NewWordlistUsecase(wordlistRepo domain.WordlistRepository, uploadDir string) WordlistUsecase {
	return &wordlistUsecase{
		wordlistRepo: wordlistRepo,
		uploadDir:    uploadDir,
	}
}

func (u *wordlistUsecase) UploadWordlist(ctx context.Context, name string, content io.Reader, size int64) (*domain.Wordlist, error) {
	// Check if wordlist with same original name already exists
	existingWordlist, err := u.wordlistRepo.GetByOrigName(ctx, name)
	if err == nil && existingWordlist != nil {
		return nil, fmt.Errorf("file already exists: %s", name)
	}

	// Create upload directory if it doesn't exist
	wordlistDir := filepath.Join(u.uploadDir, "wordlists")
	if err := os.MkdirAll(wordlistDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create wordlist directory: %w", err)
	}

	// Generate unique filename
	fileID := uuid.New()
	ext := filepath.Ext(name)
	if ext == "" {
		ext = ".txt" // Default extension for wordlists
	}
	filename := fmt.Sprintf("%s%s", fileID.String(), ext)
	filePath := filepath.Join(wordlistDir, filename)

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content to file and count words
	wordCount, written, err := u.copyAndCountWords(file, content)
	if err != nil {
		// Clean up on error
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Create wordlist record
	wordlist := &domain.Wordlist{
		ID:        fileID,
		Name:      filename,
		OrigName:  name,
		Path:      filePath,
		Size:      written,
		WordCount: &wordCount,
	}

	if err := u.wordlistRepo.Create(ctx, wordlist); err != nil {
		// Clean up on error
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to create wordlist record: %w", err)
	}

	return wordlist, nil
}

func (u *wordlistUsecase) GetWordlist(ctx context.Context, id uuid.UUID) (*domain.Wordlist, error) {
	wordlist, err := u.wordlistRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get wordlist: %w", err)
	}
	return wordlist, nil
}

func (u *wordlistUsecase) GetByOrigName(ctx context.Context, origName string) (*domain.Wordlist, error) {
	wordlist, err := u.wordlistRepo.GetByOrigName(ctx, origName)
	if err != nil {
		return nil, fmt.Errorf("failed to get wordlist by original name: %w", err)
	}
	return wordlist, nil
}

func (u *wordlistUsecase) GetAllWordlists(ctx context.Context) ([]domain.Wordlist, error) {
	wordlists, err := u.wordlistRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get wordlists: %w", err)
	}
	return wordlists, nil
}

func (u *wordlistUsecase) DeleteWordlist(ctx context.Context, id uuid.UUID) error {
	wordlist, err := u.wordlistRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get wordlist: %w", err)
	}

	// Delete the physical file
	if err := os.Remove(wordlist.Path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete physical file: %w", err)
	}

	// Delete the record
	if err := u.wordlistRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete wordlist record: %w", err)
	}

	return nil
}

func (u *wordlistUsecase) copyAndCountWords(dst io.Writer, src io.Reader) (int64, int64, error) {
	var wordCount int64 = 0
	var bytesWritten int64 = 0

	scanner := bufio.NewScanner(src)
	writer := bufio.NewWriter(dst)
	defer writer.Flush()

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line != "" { // Count non-empty lines as words
			wordCount++

			// Write line to destination
			n, err := writer.WriteString(line + "\n")
			if err != nil {
				return 0, 0, err
			}
			bytesWritten += int64(n)
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, err
	}

	return wordCount, bytesWritten, nil
}
