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

// MockWordlistUsecase is a mock implementation of usecase.WordlistUsecase
type MockWordlistUsecase struct {
	mock.Mock
}

func (m *MockWordlistUsecase) UploadWordlist(ctx context.Context, name string, content io.Reader, size int64) (*domain.Wordlist, error) {
	args := m.Called(ctx, name, content, size)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Wordlist), args.Error(1)
}

func (m *MockWordlistUsecase) GetWordlist(ctx context.Context, id uuid.UUID) (*domain.Wordlist, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Wordlist), args.Error(1)
}

func (m *MockWordlistUsecase) GetAllWordlists(ctx context.Context) ([]*domain.Wordlist, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Wordlist), args.Error(1)
}

func (m *MockWordlistUsecase) DeleteWordlist(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestWordlistHandler_UploadWordlist(t *testing.T) {
	tests := []struct {
		name           string
		setupRequest   func() (*http.Request, error)
		mockSetup      func(*MockWordlistUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful wordlist upload",
			setupRequest: func() (*http.Request, error) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				fw, err := writer.CreateFormFile("file", "rockyou.txt")
				if err != nil {
					return nil, err
				}
				_, err = io.WriteString(fw, "password\n123456\nadmin\ntest")
				if err != nil {
					return nil, err
				}
				writer.Close()

				req, err := http.NewRequest("POST", "/wordlists", body)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req, nil
			},
			mockSetup: func(mockUsecase *MockWordlistUsecase) {
				expectedWordlist := &domain.Wordlist{
					ID:       uuid.New(),
					Name:     "rockyou.txt",
					OrigName: "rockyou.txt",
					Size:     26,
				}
				mockUsecase.On("UploadWordlist", mock.Anything, "rockyou.txt", mock.Anything, mock.AnythingOfType("int64")).Return(expectedWordlist, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				data := response["data"].(map[string]interface{})
				assert.Equal(t, "rockyou.txt", data["orig_name"])
			},
		},
		{
			name: "no file provided",
			setupRequest: func() (*http.Request, error) {
				return http.NewRequest("POST", "/wordlists", strings.NewReader(""))
			},
			mockSetup: func(mockUsecase *MockWordlistUsecase) {
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
				fw, err := writer.CreateFormFile("file", "test.txt")
				if err != nil {
					return nil, err
				}
				_, err = io.WriteString(fw, "password1\npassword2")
				if err != nil {
					return nil, err
				}
				writer.Close()

				req, err := http.NewRequest("POST", "/wordlists", body)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req, nil
			},
			mockSetup: func(mockUsecase *MockWordlistUsecase) {
				mockUsecase.On("UploadWordlist", mock.Anything, "test.txt", mock.Anything, mock.AnythingOfType("int64")).Return(nil, errors.New("failed to save file"))
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
			mockUsecase := new(MockWordlistUsecase)
			tt.mockSetup(mockUsecase)

			handler := handler.NewWordlistHandler(mockUsecase)
			router := setupTestRouter()
			router.POST("/wordlists", handler.UploadWordlist)

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

func TestWordlistHandler_GetWordlist(t *testing.T) {
	wordlistID := uuid.New()

	tests := []struct {
		name           string
		wordlistID     string
		mockSetup      func(*MockWordlistUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "successful wordlist retrieval",
			wordlistID: wordlistID.String(),
			mockSetup: func(mockUsecase *MockWordlistUsecase) {
				expectedWordlist := &domain.Wordlist{
					ID:       wordlistID,
					Name:     "rockyou.txt",
					OrigName: "rockyou.txt",
					Size:     1024,
				}
				mockUsecase.On("GetWordlist", mock.Anything, wordlistID).Return(expectedWordlist, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				data := response["data"].(map[string]interface{})
				assert.Equal(t, wordlistID.String(), data["id"])
				assert.Equal(t, "rockyou.txt", data["name"])
			},
		},
		{
			name:       "invalid wordlist ID",
			wordlistID: "invalid-uuid",
			mockSetup: func(mockUsecase *MockWordlistUsecase) {
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
			name:       "wordlist not found",
			wordlistID: wordlistID.String(),
			mockSetup: func(mockUsecase *MockWordlistUsecase) {
				mockUsecase.On("GetWordlist", mock.Anything, wordlistID).Return(nil, errors.New("wordlist not found"))
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "wordlist not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockWordlistUsecase)
			tt.mockSetup(mockUsecase)

			handler := handler.NewWordlistHandler(mockUsecase)
			router := setupTestRouter()
			router.GET("/wordlists/:id", handler.GetWordlist)

			req, err := http.NewRequest("GET", "/wordlists/"+tt.wordlistID, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestWordlistHandler_GetAllWordlists(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockWordlistUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful wordlists retrieval",
			mockSetup: func(mockUsecase *MockWordlistUsecase) {
				wordlists := []*domain.Wordlist{
					{
						ID:       uuid.New(),
						Name:     "rockyou.txt",
						OrigName: "rockyou.txt",
						Size:     1024,
					},
					{
						ID:       uuid.New(),
						Name:     "common.txt",
						OrigName: "common-passwords.txt",
						Size:     512,
					},
				}
				mockUsecase.On("GetAllWordlists", mock.Anything).Return(wordlists, nil)
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
			name: "no wordlists found",
			mockSetup: func(mockUsecase *MockWordlistUsecase) {
				mockUsecase.On("GetAllWordlists", mock.Anything).Return([]*domain.Wordlist{}, nil)
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
			mockSetup: func(mockUsecase *MockWordlistUsecase) {
				mockUsecase.On("GetAllWordlists", mock.Anything).Return([]*domain.Wordlist{}, errors.New("database error"))
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
			mockUsecase := new(MockWordlistUsecase)
			tt.mockSetup(mockUsecase)

			handler := handler.NewWordlistHandler(mockUsecase)
			router := setupTestRouter()
			router.GET("/wordlists", handler.GetAllWordlists)

			req, err := http.NewRequest("GET", "/wordlists", nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestWordlistHandler_DeleteWordlist(t *testing.T) {
	wordlistID := uuid.New()

	tests := []struct {
		name           string
		wordlistID     string
		mockSetup      func(*MockWordlistUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "successful wordlist deletion",
			wordlistID: wordlistID.String(),
			mockSetup: func(mockUsecase *MockWordlistUsecase) {
				mockUsecase.On("DeleteWordlist", mock.Anything, wordlistID).Return(nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Wordlist deleted successfully", response["message"])
			},
		},
		{
			name:       "invalid wordlist ID",
			wordlistID: "invalid-uuid",
			mockSetup: func(mockUsecase *MockWordlistUsecase) {
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
			name:       "wordlist not found",
			wordlistID: wordlistID.String(),
			mockSetup: func(mockUsecase *MockWordlistUsecase) {
				mockUsecase.On("DeleteWordlist", mock.Anything, wordlistID).Return(errors.New("wordlist not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "wordlist not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockWordlistUsecase)
			tt.mockSetup(mockUsecase)

			handler := handler.NewWordlistHandler(mockUsecase)
			router := setupTestRouter()
			router.DELETE("/wordlists/:id", handler.DeleteWordlist)

			req, err := http.NewRequest("DELETE", "/wordlists/"+tt.wordlistID, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockUsecase.AssertExpectations(t)
		})
	}
}
