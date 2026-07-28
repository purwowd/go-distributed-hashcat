package usecase

import (
	"bufio"
	"context"
	"encoding/json"
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
	RegisterLocalWordlist(ctx context.Context, req *domain.RegisterLocalWordlistRequest) (*domain.Wordlist, error)
	SyncAgentLocalWordlists(ctx context.Context, files []domain.AgentLocalFile) error
	GetWordlist(ctx context.Context, id uuid.UUID) (*domain.Wordlist, error)
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
		Source:    domain.WordlistSourceUploaded,
	}

	if err := u.wordlistRepo.Create(ctx, wordlist); err != nil {
		// Clean up on error
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to create wordlist record: %w", err)
	}

	return wordlist, nil
}

func normalizeWordlistOrigName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	if !strings.HasSuffix(strings.ToLower(name), ".txt") {
		name += ".txt"
	}
	return name
}

func agentLocalWordlistPath(origName string) string {
	return "agent_local:" + origName
}

func (u *wordlistUsecase) RegisterLocalWordlist(ctx context.Context, req *domain.RegisterLocalWordlistRequest) (*domain.Wordlist, error) {
	origName := normalizeWordlistOrigName(req.OrigName)
	if origName == "" {
		return nil, fmt.Errorf("orig_name is required")
	}

	if existing, err := u.wordlistRepo.GetByOrigName(ctx, origName); err == nil && existing != nil {
		if existing.Source == domain.WordlistSourceUploaded {
			return existing, nil
		}
		if req.Size > existing.Size {
			existing.Size = req.Size
		}
		if req.WordCount > 0 {
			existing.WordCount = &req.WordCount
		}
		if err := u.wordlistRepo.Update(ctx, existing); err != nil {
			return nil, fmt.Errorf("failed to update local wordlist metadata: %w", err)
		}
		return existing, nil
	}

	wordCount := req.WordCount
	wordlist := &domain.Wordlist{
		ID:        uuid.New(),
		Name:      origName,
		OrigName:  origName,
		Path:      agentLocalWordlistPath(origName),
		Size:      req.Size,
		WordCount: &wordCount,
		Source:    domain.WordlistSourceAgentLocal,
	}

	if err := u.wordlistRepo.Create(ctx, wordlist); err != nil {
		return nil, fmt.Errorf("failed to register local wordlist: %w", err)
	}
	return wordlist, nil
}

func (u *wordlistUsecase) SyncAgentLocalWordlists(ctx context.Context, files []domain.AgentLocalFile) error {
	// Companion *.txt.meta.json files (written beside wordlists) carry word_count.
	metaCounts := make(map[string]int64)
	for _, file := range files {
		lowerName := strings.ToLower(file.Name)
		if !strings.HasSuffix(lowerName, ".meta.json") {
			continue
		}
		baseName := file.Name[:len(file.Name)-len(".meta.json")]
		if !strings.HasSuffix(strings.ToLower(baseName), ".txt") {
			continue
		}
		if wc, ok := readWordlistMetaWordCount(file.Path); ok && wc > 0 {
			metaCounts[baseName] = wc
		}
	}

	for _, file := range files {
		if file.Type != "" && file.Type != "wordlist" {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(file.Name), ".txt") {
			continue
		}
		wordCount := metaCounts[file.Name]
		if wordCount == 0 {
			if wc, ok := readWordlistMetaWordCount(file.Path + ".meta.json"); ok {
				wordCount = wc
			}
		}
		_, err := u.RegisterLocalWordlist(ctx, &domain.RegisterLocalWordlistRequest{
			OrigName:  file.Name,
			Size:      file.Size,
			WordCount: wordCount,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

type wordlistMetaFile struct {
	WordCount int64 `json:"word_count"`
	FileSize  int64 `json:"file_size"`
}

func readWordlistMetaWordCount(metaPath string) (int64, bool) {
	if metaPath == "" {
		return 0, false
	}
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return 0, false
	}
	var meta wordlistMetaFile
	if err := json.Unmarshal(data, &meta); err != nil {
		return 0, false
	}
	if meta.WordCount <= 0 {
		return 0, false
	}
	return meta.WordCount, true
}

func (u *wordlistUsecase) GetWordlist(ctx context.Context, id uuid.UUID) (*domain.Wordlist, error) {
	wordlist, err := u.wordlistRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get wordlist: %w", err)
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

	// Delete the physical file only for uploaded wordlists stored on the server.
	if wordlist.Source != domain.WordlistSourceAgentLocal && !strings.HasPrefix(wordlist.Path, "agent_local:") {
		if err := os.Remove(wordlist.Path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete physical file: %w", err)
		}
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
