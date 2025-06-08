package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-distributed-hashcat/internal/delivery/http/handler"
	"go-distributed-hashcat/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHashFileUsecase is a mock implementation of usecase.HashFileUsecase
type MockHashFileUsecase struct {
	mock.Mock
}

func (m *MockHashFileUsecase) UploadHashFile(ctx context.Context, name string, content io.Reader, size int64) (*domain.HashFile, error) {
	args := m.Called(ctx, name, content, size)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.HashFile), args.Error(1)
}

func (m *MockHashFileUsecase) GetHashFile(ctx context.Context, id uuid.UUID) (*domain.HashFile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.HashFile), args.Error(1)
}

func (m *MockHashFileUsecase) GetAllHashFiles(ctx context.Context) ([]domain.HashFile, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.HashFile), args.Error(1)
}

func (m *MockHashFileUsecase) DeleteHashFile(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestHashFileHandler_UploadHashFile(t *testing.T) {
	tests := []struct {
		name           string
		setupRequest   func() (*http.Request, error)
		mockSetup      func(*MockHashFileUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful hash file upload",
			setupRequest: func() (*http.Request, error) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				fw, err := writer.CreateFormFile("file", "test.hash")
				if err != nil {
					return nil, err
				}
				_, err = io.WriteString(fw, "5d41402abc4b2a76b9719d911017c592\n")
				if err != nil {
					return nil, err
				}
				writer.Close()

				req, err := http.NewRequest("POST", "/hashfiles", body)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req, nil
			},
			mockSetup: func(mockUsecase *MockHashFileUsecase) {
				expectedHashFile := &domain.HashFile{
					ID:       uuid.New(),
					Name:     "test.hash",
					OrigName: "test.hash",
					Size:     33,
				}
				mockUsecase.On("UploadHashFile", mock.Anything, "test.hash", mock.Anything, mock.AnythingOfType("int64")).Return(expectedHashFile, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				data := response["data"].(map[string]interface{})
				assert.Equal(t, "test.hash", data["orig_name"])
			},
		},
		{
			name: "no file provided",
			setupRequest: func() (*http.Request, error) {
				return http.NewRequest("POST", "/hashfiles", strings.NewReader(""))
			},
			mockSetup: func(mockUsecase *MockHashFileUsecase) {
				// No mock calls expected
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "No file")
			},
		},
		{
			name: "usecase error",
			setupRequest: func() (*http.Request, error) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				fw, err := writer.CreateFormFile("file", "test.hash")
				if err != nil {
					return nil, err
				}
				_, err = io.WriteString(fw, "test content")
				if err != nil {
					return nil, err
				}
				writer.Close()

				req, err := http.NewRequest("POST", "/hashfiles", body)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req, nil
			},
			mockSetup: func(mockUsecase *MockHashFileUsecase) {
				mockUsecase.On("UploadHashFile", mock.Anything, "test.hash", mock.Anything, mock.AnythingOfType("int64")).Return(nil, errors.New("failed to save file"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "failed to save file")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockHashFileUsecase)
			tt.mockSetup(mockUsecase)

			handler := handler.NewHashFileHandler(mockUsecase)
			router := setupTestRouter()
			router.POST("/hashfiles", handler.UploadHashFile)

			req, err := tt.setupRequest()
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestHashFileHandler_GetHashFile(t *testing.T) {
	hashFileID := uuid.New()

	tests := []struct {
		name           string
		hashFileID     string
		mockSetup      func(*MockHashFileUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "successful hash file retrieval",
			hashFileID: hashFileID.String(),
			mockSetup: func(mockUsecase *MockHashFileUsecase) {
				expectedHashFile := &domain.HashFile{
					ID:       hashFileID,
					Name:     "test.hash",
					OrigName: "test.hash",
					Size:     1024,
				}
				mockUsecase.On("GetHashFile", mock.Anything, hashFileID).Return(expectedHashFile, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				data := response["data"].(map[string]interface{})
				assert.Equal(t, hashFileID.String(), data["id"])
				assert.Equal(t, "test.hash", data["name"])
			},
		},
		{
			name:       "invalid hash file ID",
			hashFileID: "invalid-uuid",
			mockSetup: func(mockUsecase *MockHashFileUsecase) {
				// No mock calls expected
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid")
			},
		},
		{
			name:       "hash file not found",
			hashFileID: hashFileID.String(),
			mockSetup: func(mockUsecase *MockHashFileUsecase) {
				mockUsecase.On("GetHashFile", mock.Anything, hashFileID).Return(nil, errors.New("hash file not found"))
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "hash file not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockHashFileUsecase)
			tt.mockSetup(mockUsecase)

			handler := handler.NewHashFileHandler(mockUsecase)
			router := setupTestRouter()
			router.GET("/hashfiles/:id", handler.GetHashFile)

			req, err := http.NewRequest("GET", "/hashfiles/"+tt.hashFileID, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestHashFileHandler_GetAllHashFiles(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockHashFileUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful hash files retrieval",
			mockSetup: func(mockUsecase *MockHashFileUsecase) {
				hashFiles := []domain.HashFile{
					{
						ID:       uuid.New(),
						Name:     "file1.hash",
						OrigName: "file1.hash",
						Size:     1024,
					},
					{
						ID:       uuid.New(),
						Name:     "file2.hash",
						OrigName: "file2.hash",
						Size:     2048,
					},
				}
				mockUsecase.On("GetAllHashFiles", mock.Anything).Return(hashFiles, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				data := response["data"].([]interface{})
				assert.Len(t, data, 2)
			},
		},
		{
			name: "no hash files found",
			mockSetup: func(mockUsecase *MockHashFileUsecase) {
				mockUsecase.On("GetAllHashFiles", mock.Anything).Return([]domain.HashFile{}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				data := response["data"].([]interface{})
				assert.Len(t, data, 0)
			},
		},
		{
			name: "usecase error",
			mockSetup: func(mockUsecase *MockHashFileUsecase) {
				mockUsecase.On("GetAllHashFiles", mock.Anything).Return([]domain.HashFile{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "database error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockHashFileUsecase)
			tt.mockSetup(mockUsecase)

			handler := handler.NewHashFileHandler(mockUsecase)
			router := setupTestRouter()
			router.GET("/hashfiles", handler.GetAllHashFiles)

			req, err := http.NewRequest("GET", "/hashfiles", nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestHashFileHandler_DeleteHashFile(t *testing.T) {
	hashFileID := uuid.New()

	tests := []struct {
		name           string
		hashFileID     string
		mockSetup      func(*MockHashFileUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "successful hash file deletion",
			hashFileID: hashFileID.String(),
			mockSetup: func(mockUsecase *MockHashFileUsecase) {
				mockUsecase.On("DeleteHashFile", mock.Anything, hashFileID).Return(nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Hash file deleted successfully", response["message"])
			},
		},
		{
			name:       "invalid hash file ID",
			hashFileID: "invalid-uuid",
			mockSetup: func(mockUsecase *MockHashFileUsecase) {
				// No mock calls expected
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid")
			},
		},
		{
			name:       "hash file not found",
			hashFileID: hashFileID.String(),
			mockSetup: func(mockUsecase *MockHashFileUsecase) {
				mockUsecase.On("DeleteHashFile", mock.Anything, hashFileID).Return(errors.New("hash file not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "hash file not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockHashFileUsecase)
			tt.mockSetup(mockUsecase)

			handler := handler.NewHashFileHandler(mockUsecase)
			router := setupTestRouter()
			router.DELETE("/hashfiles/:id", handler.DeleteHashFile)

			req, err := http.NewRequest("DELETE", "/hashfiles/"+tt.hashFileID, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockUsecase.AssertExpectations(t)
		})
	}
}
