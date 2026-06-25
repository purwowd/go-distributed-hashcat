package benchmarks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go-distributed-hashcat/internal/delivery/http/handler"
	"go-distributed-hashcat/internal/infrastructure/database"
	"go-distributed-hashcat/internal/infrastructure/repository"
	"go-distributed-hashcat/internal/usecase"

	"github.com/gin-gonic/gin"
)

func TestJobsAPIResponseSize(t *testing.T) {
	if os.Getenv("RUN_DB_BENCH") != "1" {
		t.Skip("set RUN_DB_BENCH=1 to run against local data/hashcat.db")
	}

	gin.SetMode(gin.TestMode)
	db, err := database.NewSQLiteDB("../../data/hashcat.db")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()

	jobRepo := repository.NewJobRepository(db)
	agentRepo := repository.NewAgentRepository(db)
	wordlistRepo := repository.NewWordlistRepository(db)
	hashFileRepo := repository.NewHashFileRepository(db)

	jobUsecase := usecase.NewJobUsecase(jobRepo, agentRepo, hashFileRepo, wordlistRepo, nil)
	enrichment := usecase.NewJobEnrichmentService(agentRepo, wordlistRepo, hashFileRepo)
	jobHandler := handler.NewJobHandler(jobUsecase, enrichment, usecase.NewAgentUsecase(agentRepo, nil), usecase.NewWordlistUsecase(wordlistRepo, "uploads/wordlists"))

	router := gin.New()
	router.GET("/jobs", jobHandler.GetAllJobs)

	req := httptest.NewRequest(http.MethodGet, "/jobs?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d body=%s", w.Code, w.Body.String())
	}

	body := w.Body.Bytes()
	t.Logf("response bytes: %d (%.2f MB)", len(body), float64(len(body))/1024/1024)

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("json: %v", err)
	}

	data, _ := payload["data"].([]interface{})
	t.Logf("items: %d total: %v page: %v", len(data), payload["total"], payload["page"])

	if len(data) > 0 {
		first, _ := data[0].(map[string]interface{})
		for _, key := range []string{"hash_file", "wordlist", "rules", "result", "hash_file_name", "wordlist_name"} {
			if v, ok := first[key]; ok {
				t.Logf("field %s len=%d", key, len([]byte(toString(v))))
			}
		}
	}

	ctx := context.Background()
	jobs, total, err := jobRepo.GetPaginated(ctx, 1, 10, "", "")
	if err != nil {
		t.Fatalf("repo paginated: %v", err)
	}
	t.Logf("repo paginated: %d jobs total=%d", len(jobs), total)
}

func toString(v interface{}) string {
	s, _ := v.(string)
	return s
}
