package workflowstep

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupRouter は実際のGin、GORM、インメモリSQLiteを使うHTTPテスト境界を準備します。
func setupRouter(t *testing.T) (*gin.Engine, *Repository) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("テストDBを開けません: %v", err)
	}
	if err := db.AutoMigrate(&WorkflowStep{}); err != nil {
		t.Fatalf("テストDBをマイグレーションできません: %v", err)
	}

	repository := NewRepository(db)
	handler := NewHandler(repository)
	router := gin.New()
	router.POST("/workflow-steps", handler.CreateWorkflowStep)
	return router, repository
}

func TestCreateWorkflowStep_DuplicateOrderMustBeRejected(t *testing.T) {
	router, repository := setupRouter(t)
	if err := repository.Create(&WorkflowStep{
		ID:         "step-existing-order",
		WorkflowID: "approval-flow",
		Name:       "申請",
		Order:      1,
	}); err != nil {
		t.Fatalf("既存ステップを作成できません: %v", err)
	}

	payload := WorkflowStep{
		WorkflowID:  "approval-flow",
		Name:        "承認",
		Description: "既存ステップと同じ順序",
		Order:       1,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("リクエストをJSON化できません: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/workflow-steps", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Errorf("status: expected %d, actual %d", http.StatusConflict, response.Code)
	}

	count, err := repository.CountByWorkflowAndOrder("approval-flow", 1)
	if err != nil {
		t.Fatalf("最終状態を読み直せません: %v", err)
	}
	if count != 1 {
		t.Errorf("persisted count: expected 1, actual %d", count)
	}
}

func TestCreateWorkflowStep_SameOrderAcrossWorkflowsIsAllowed(t *testing.T) {
	router, repository := setupRouter(t)
	if err := repository.Create(&WorkflowStep{
		ID:         "step-other-flow",
		WorkflowID: "other-flow",
		Name:       "申請",
		Order:      1,
	}); err != nil {
		t.Fatalf("別ワークフローの既存ステップを作成できません: %v", err)
	}

	payload := WorkflowStep{
		WorkflowID: "approval-flow",
		Name:       "申請",
		Order:      1,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("リクエストをJSON化できません: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/workflow-steps", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Errorf("status: expected %d, actual %d", http.StatusCreated, response.Code)
	}
}
