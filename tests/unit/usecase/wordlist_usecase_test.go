package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"go-distributed-hashcat/internal/domain"
	"go-distributed-hashcat/internal/usecase"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWordlistRepository for wordlist tests
type MockWordlistRepository struct {
	mock.Mock
}

func (m *MockWordlistRepository) Create(ctx context.Context, wordlist *domain.Wordlist) error {
	args := m.Called(ctx, wordlist)
	return args.Error(0)
}

func (m *MockWordlistRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Wordlist, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Wordlist), args.Error(1)
}

func (m *MockWordlistRepository) GetByOrigName(ctx context.Context, origName string) (*domain.Wordlist, error) {
	args := m.Called(ctx, origName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Wordlist), args.Error(1)
}

func (m *MockWordlistRepository) GetAll(ctx context.Context) ([]domain.Wordlist, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Wordlist), args.Error(1)
}

func (m *MockWordlistRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestWordlistUsecase_UploadWordlist(t *testing.T) {
	tests := []struct {
		name          string
		filename      string
		fileContent   string
		mockSetup     func(*MockWordlistRepository)
		expectedError bool
	}{
		{
			name: "successful wordlist upload",
			mockSetup: func(repo *MockWordlistRepository) {
				repo.On("GetByOrigName", mock.Anything, "test.txt").Return(nil, fmt.Errorf("wordlist not found"))
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Wordlist")).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "duplicate file name",
			mockSetup: func(repo *MockWordlistRepository) {
				existingWordlist := &domain.Wordlist{
					ID:       uuid.New(),
					Name:     "existing.txt",
					OrigName: "test.txt",
					Path:     "/tmp/wordlists/existing.txt",
					Size:     1024,
				}
				repo.On("GetByOrigName", mock.Anything, "test.txt").Return(existingWordlist, nil)
			},
			expectedError: true,
		},
		{
			name:        "repository error during creation",
			filename:    "test.txt",
			fileContent: "password1\npassword2",
			mockSetup: func(repo *MockWordlistRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Wordlist")).Return(errors.New("database error"))
			},
			expectedError: true,
		},
		{
			name:        "empty wordlist",
			filename:    "empty.txt",
			fileContent: "",
			mockSetup: func(repo *MockWordlistRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Wordlist")).Return(nil)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockWordlistRepository)
			tt.mockSetup(mockRepo)

			usecase := usecase.NewWordlistUsecase(mockRepo, "/tmp/wordlists")
			ctx := context.Background()

			wordlist, err := usecase.UploadWordlist(ctx, tt.filename, strings.NewReader(tt.fileContent), int64(len(tt.fileContent)))

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, wordlist)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, wordlist)
				assert.Equal(t, tt.filename, wordlist.OrigName)
				assert.NotEqual(t, uuid.Nil, wordlist.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWordlistUsecase_GetWordlist(t *testing.T) {
	wordlistID := uuid.New()
	expectedWordlist := &domain.Wordlist{
		ID:       wordlistID,
		Name:     "rockyou.txt",
		OrigName: "rockyou.txt",
		Path:     "/tmp/wordlists/rockyou.txt",
		Size:     1024,
	}

	tests := []struct {
		name          string
		wordlistID    uuid.UUID
		mockSetup     func(*MockWordlistRepository)
		expectedError bool
	}{
		{
			name:       "successful wordlist retrieval",
			wordlistID: wordlistID,
			mockSetup: func(repo *MockWordlistRepository) {
				repo.On("GetByID", mock.Anything, wordlistID).Return(expectedWordlist, nil)
			},
			expectedError: false,
		},
		{
			name:       "wordlist not found",
			wordlistID: wordlistID,
			mockSetup: func(repo *MockWordlistRepository) {
				repo.On("GetByID", mock.Anything, wordlistID).Return(nil, errors.New("wordlist not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockWordlistRepository)
			tt.mockSetup(mockRepo)

			usecase := usecase.NewWordlistUsecase(mockRepo, "/tmp/wordlists")
			ctx := context.Background()

			wordlist, err := usecase.GetWordlist(ctx, tt.wordlistID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, wordlist)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, wordlist)
				assert.Equal(t, expectedWordlist.ID, wordlist.ID)
				assert.Equal(t, expectedWordlist.Name, wordlist.Name)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWordlistUsecase_GetByOrigName(t *testing.T) {
	expectedWordlist := &domain.Wordlist{
		ID:       uuid.New(),
		Name:     "rockyou.txt",
		OrigName: "rockyou.txt",
		Path:     "/tmp/wordlists/rockyou.txt",
		Size:     1024,
	}

	tests := []struct {
		name          string
		origName      string
		mockSetup     func(*MockWordlistRepository)
		expectedError bool
	}{
		{
			name:     "successful wordlist retrieval by original name",
			origName: "rockyou.txt",
			mockSetup: func(repo *MockWordlistRepository) {
				repo.On("GetByOrigName", mock.Anything, "rockyou.txt").Return(expectedWordlist, nil)
			},
			expectedError: false,
		},
		{
			name:     "wordlist not found by original name",
			origName: "nonexistent.txt",
			mockSetup: func(repo *MockWordlistRepository) {
				repo.On("GetByOrigName", mock.Anything, "nonexistent.txt").Return(nil, fmt.Errorf("wordlist not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockWordlistRepository)
			tt.mockSetup(mockRepo)

			usecase := usecase.NewWordlistUsecase(mockRepo, "/tmp/wordlists")
			ctx := context.Background()

			wordlist, err := usecase.GetByOrigName(ctx, tt.origName)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedWordlist, wordlist)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWordlistUsecase_GetAllWordlists(t *testing.T) {
	expectedWordlists := []domain.Wordlist{
		{
			ID:       uuid.New(),
			Name:     "rockyou.txt",
			OrigName: "rockyou.txt",
			Path:     "/tmp/wordlists/rockyou.txt",
			Size:     1024,
		},
		{
			ID:       uuid.New(),
			Name:     "common.txt",
			OrigName: "common-passwords.txt",
			Path:     "/tmp/wordlists/common.txt",
			Size:     512,
		},
	}

	tests := []struct {
		name          string
		mockSetup     func(*MockWordlistRepository)
		expectedCount int
		expectedError bool
	}{
		{
			name: "successful wordlists retrieval",
			mockSetup: func(repo *MockWordlistRepository) {
				repo.On("GetAll", mock.Anything).Return(expectedWordlists, nil)
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "no wordlists found",
			mockSetup: func(repo *MockWordlistRepository) {
				repo.On("GetAll", mock.Anything).Return([]domain.Wordlist{}, nil)
			},
			expectedCount: 0,
			expectedError: false,
		},
		{
			name: "repository error",
			mockSetup: func(repo *MockWordlistRepository) {
				repo.On("GetAll", mock.Anything).Return([]domain.Wordlist{}, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockWordlistRepository)
			tt.mockSetup(mockRepo)

			usecase := usecase.NewWordlistUsecase(mockRepo, "/tmp/wordlists")
			ctx := context.Background()

			wordlists, err := usecase.GetAllWordlists(ctx)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, wordlists, tt.expectedCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWordlistUsecase_DeleteWordlist(t *testing.T) {
	wordlistID := uuid.New()
	expectedWordlist := &domain.Wordlist{
		ID:       wordlistID,
		Name:     "test.txt",
		OrigName: "test-wordlist.txt",
		Path:     "/tmp/wordlists/test.txt",
		Size:     1024,
	}

	tests := []struct {
		name          string
		wordlistID    uuid.UUID
		mockSetup     func(*MockWordlistRepository)
		expectedError bool
	}{
		{
			name:       "successful wordlist deletion",
			wordlistID: wordlistID,
			mockSetup: func(repo *MockWordlistRepository) {
				repo.On("GetByID", mock.Anything, wordlistID).Return(expectedWordlist, nil)
				repo.On("Delete", mock.Anything, wordlistID).Return(nil)
			},
			expectedError: false,
		},
		{
			name:       "wordlist not found",
			wordlistID: wordlistID,
			mockSetup: func(repo *MockWordlistRepository) {
				repo.On("GetByID", mock.Anything, wordlistID).Return(nil, errors.New("wordlist not found"))
			},
			expectedError: true,
		},
		{
			name:       "repository error during deletion",
			wordlistID: wordlistID,
			mockSetup: func(repo *MockWordlistRepository) {
				repo.On("GetByID", mock.Anything, wordlistID).Return(expectedWordlist, nil)
				repo.On("Delete", mock.Anything, wordlistID).Return(errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockWordlistRepository)
			tt.mockSetup(mockRepo)

			usecase := usecase.NewWordlistUsecase(mockRepo, "/tmp/wordlists")
			ctx := context.Background()

			err := usecase.DeleteWordlist(ctx, tt.wordlistID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
