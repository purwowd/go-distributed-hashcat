package usecase

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"go-distributed-hashcat/internal/domain"

	"github.com/google/uuid"
)

// formatBytes converts bytes to human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

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

	// Create the file with buffered I/O for better performance
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Use buffered writer for better performance
	bufferedFile := bufio.NewWriterSize(file, 64*1024) // 64KB buffer
	defer bufferedFile.Flush()

	// Copy content to file with progress tracking for large files
	var wordCount int64
	var written int64
	
	if size > 100*1024*1024 { // > 100MB
		// For large files, use progress tracking
		wordCount, written, err = u.copyAndCountWordsWithProgress(bufferedFile, content, size)
	} else {
		// For smaller files, use regular method
		wordCount, written, err = u.copyAndCountWords(bufferedFile, content)
	}
	
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

// copyAndCountWordsWithProgress handles large files with progress tracking
func (u *wordlistUsecase) copyAndCountWordsWithProgress(writer io.Writer, reader io.Reader, totalSize int64) (int64, int64, error) {
	var wordCount int64
	var written int64
	buffer := make([]byte, 32*1024) // 32KB buffer for reading
	
	// Progress tracking variables
	lastProgress := int64(0)
	progressInterval := int64(10 * 1024 * 1024) // Log every 10MB
	
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			// Write to file
			wn, werr := writer.Write(buffer[:n])
			if werr != nil {
				return wordCount, written, werr
			}
			written += int64(wn)
			
			// Count words in this buffer
			wordCount += int64(bytes.Count(buffer[:n], []byte{'\n'}))
			
			// Progress tracking
			if written-lastProgress >= progressInterval {
				progress := float64(written) / float64(totalSize) * 100
				log.Printf("Upload progress: %s / %s (%.1f%%)", 
					formatBytes(written), 
					formatBytes(totalSize), 
					progress)
				lastProgress = written
			}
		}
		
		if err == io.EOF {
			break
		}
		if err != nil {
			return wordCount, written, err
		}
	}
	
	log.Printf("Upload completed: %s written, %d words counted", formatBytes(written), wordCount)
	return wordCount, written, nil
}
