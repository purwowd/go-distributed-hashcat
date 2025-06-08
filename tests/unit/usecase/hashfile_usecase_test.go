package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go-distributed-hashcat/internal/domain"
	"go-distributed-hashcat/internal/usecase"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHashFileUsecase_UploadHashFile(t *testing.T) {
	tests := []struct {
		name          string
		filename      string
		fileContent   string
		mockSetup     func(*MockHashFileRepository)
		expectedError bool
	}{
		{
			name:        "successful file upload",
			filename:    "test.hash",
			fileContent: "5d41402abc4b2a76b9719d911017c592\n8b1a9953c4611296a827abf8c47804d7",
			mockSetup: func(repo *MockHashFileRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.HashFile")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:        "repository error during creation",
			filename:    "test.hash",
			fileContent: "5d41402abc4b2a76b9719d911017c592",
			mockSetup: func(repo *MockHashFileRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.HashFile")).Return(errors.New("database error"))
			},
			expectedError: true,
		},
		{
			name:        "empty file",
			filename:    "empty.hash",
			fileContent: "",
			mockSetup: func(repo *MockHashFileRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.HashFile")).Return(nil)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockHashFileRepository)
			tt.mockSetup(mockRepo)

			usecase := usecase.NewHashFileUsecase(mockRepo, "/tmp/uploads")
			ctx := context.Background()

			hashFile, err := usecase.UploadHashFile(ctx, tt.filename, strings.NewReader(tt.fileContent), int64(len(tt.fileContent)))

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, hashFile)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, hashFile)
				assert.Equal(t, tt.filename, hashFile.OrigName)
				assert.NotEqual(t, uuid.Nil, hashFile.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestHashFileUsecase_GetHashFile(t *testing.T) {
	hashFileID := uuid.New()
	expectedHashFile := &domain.HashFile{
		ID:   hashFileID,
		Name: "test.hash",
		Path: "/uploads/test.hash",
		Size: 1024,
	}

	tests := []struct {
		name          string
		hashFileID    uuid.UUID
		mockSetup     func(*MockHashFileRepository)
		expectedError bool
	}{
		{
			name:       "successful hash file retrieval",
			hashFileID: hashFileID,
			mockSetup: func(repo *MockHashFileRepository) {
				repo.On("GetByID", mock.Anything, hashFileID).Return(expectedHashFile, nil)
			},
			expectedError: false,
		},
		{
			name:       "hash file not found",
			hashFileID: hashFileID,
			mockSetup: func(repo *MockHashFileRepository) {
				repo.On("GetByID", mock.Anything, hashFileID).Return(nil, errors.New("hash file not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockHashFileRepository)
			tt.mockSetup(mockRepo)

			usecase := usecase.NewHashFileUsecase(mockRepo, "/tmp/uploads")
			ctx := context.Background()

			hashFile, err := usecase.GetHashFile(ctx, tt.hashFileID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, hashFile)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, hashFile)
				assert.Equal(t, expectedHashFile.ID, hashFile.ID)
				assert.Equal(t, expectedHashFile.Name, hashFile.Name)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestHashFileUsecase_GetAllHashFiles(t *testing.T) {
	expectedHashFiles := []domain.HashFile{
		{
			ID:   uuid.New(),
			Name: "test1.hash",
			Path: "/uploads/test1.hash",
			Size: 1024,
		},
		{
			ID:   uuid.New(),
			Name: "test2.hash",
			Path: "/uploads/test2.hash",
			Size: 2048,
		},
	}

	tests := []struct {
		name          string
		mockSetup     func(*MockHashFileRepository)
		expectedCount int
		expectedError bool
	}{
		{
			name: "successful hash files retrieval",
			mockSetup: func(repo *MockHashFileRepository) {
				repo.On("GetAll", mock.Anything).Return(expectedHashFiles, nil)
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "no hash files found",
			mockSetup: func(repo *MockHashFileRepository) {
				repo.On("GetAll", mock.Anything).Return([]domain.HashFile{}, nil)
			},
			expectedCount: 0,
			expectedError: false,
		},
		{
			name: "repository error",
			mockSetup: func(repo *MockHashFileRepository) {
				repo.On("GetAll", mock.Anything).Return([]domain.HashFile{}, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockHashFileRepository)
			tt.mockSetup(mockRepo)

			usecase := usecase.NewHashFileUsecase(mockRepo, "/tmp/uploads")
			ctx := context.Background()

			hashFiles, err := usecase.GetAllHashFiles(ctx)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, hashFiles, tt.expectedCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestHashFileUsecase_DeleteHashFile(t *testing.T) {
	hashFileID := uuid.New()
	expectedHashFile := &domain.HashFile{
		ID:   hashFileID,
		Name: "test.hash",
		Path: "/tmp/uploads/test.hash",
		Size: 1024,
	}

	tests := []struct {
		name          string
		hashFileID    uuid.UUID
		mockSetup     func(*MockHashFileRepository)
		expectedError bool
	}{
		{
			name:       "successful hash file deletion",
			hashFileID: hashFileID,
			mockSetup: func(repo *MockHashFileRepository) {
				repo.On("GetByID", mock.Anything, hashFileID).Return(expectedHashFile, nil)
				repo.On("Delete", mock.Anything, hashFileID).Return(nil)
			},
			expectedError: false,
		},
		{
			name:       "hash file not found",
			hashFileID: hashFileID,
			mockSetup: func(repo *MockHashFileRepository) {
				repo.On("GetByID", mock.Anything, hashFileID).Return(nil, errors.New("hash file not found"))
			},
			expectedError: true,
		},
		{
			name:       "repository error during deletion",
			hashFileID: hashFileID,
			mockSetup: func(repo *MockHashFileRepository) {
				repo.On("GetByID", mock.Anything, hashFileID).Return(expectedHashFile, nil)
				repo.On("Delete", mock.Anything, hashFileID).Return(errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockHashFileRepository)
			tt.mockSetup(mockRepo)

			usecase := usecase.NewHashFileUsecase(mockRepo, "/tmp/uploads")
			ctx := context.Background()

			err := usecase.DeleteHashFile(ctx, tt.hashFileID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
