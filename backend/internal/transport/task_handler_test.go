// backend/internal/transport/task_handler_test.go func(

package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"todo_axios_api/backend/internal/models"
)

func newTestTransportDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite test db: %v", err)
	}

	if err := db.AutoMigrate(&models.Task{}); err != nil {
		t.Fatalf("failed to migrate Task model: %v", err)
	}

	return db
}

func newTestRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	api := r.Group("/api")
	{
		api.GET("/tasks", makeGetTasks(db))
		api.POST("/tasks", makeCreateTask(db))
		api.PATCH("/tasks/:id", makeUpdateTask(db))
		api.DELETE("/tasks/:id", makeDeleteTask(db))
	}

	return r
}

func performJSONRequest(r http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}

	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func seedTask(t *testing.T, db *gorm.DB, title string, dueDate time.Time, state models.TaskState) models.Task {
	t.Helper()

	task := models.Task{
		Title:       title,
		Description: title + " description",
		DueDate:     dueDate,
		State:       state,
	}

	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}

	return task
}

func TestMakeGetTasks_Empty(t *testing.T) {
	db := newTestTransportDB(t)
	router := newTestRouter(db)

	w := performJSONRequest(router, http.MethodGet, "/api/tasks", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d; body=%s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp struct {
		Tasks []models.Task `json:"tasks"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Tasks) != 0 {
		t.Fatalf("expected 0 tasks, got %d", len(resp.Tasks))
	}
}

func TestMakeGetTasks_OrdersByDueDateAsc(t *testing.T) {
	db := newTestTransportDB(t)
	router := newTestRouter(db)

	later := seedTask(t, db, "Later", time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC), models.Scheduled)
	earlier := seedTask(t, db, "Earlier", time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC), models.Scheduled)

	w := performJSONRequest(router, http.MethodGet, "/api/tasks", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d; body=%s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp struct {
		Tasks []models.Task `json:"tasks"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(resp.Tasks))
	}

	if resp.Tasks[0].ID != earlier.ID {
		t.Fatalf("expected first task ID %d, got %d", earlier.ID, resp.Tasks[0].ID)
	}
	if resp.Tasks[1].ID != later.ID {
		t.Fatalf("expected second task ID %d, got %d", later.ID, resp.Tasks[1].ID)
	}
}

func TestMakeCreateTask_Success(t *testing.T) {
	db := newTestTransportDB(t)
	router := newTestRouter(db)

	body := map[string]any{
		"title":       "Pay bills",
		"description": "Electric + water",
		"dueDate":     "2026-03-07",
	}

	w := performJSONRequest(router, http.MethodPost, "/api/tasks", body)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d; body=%s", http.StatusCreated, w.Code, w.Body.String())
	}

	var resp struct {
		Task models.Task `json:"task"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Task.ID == 0 {
		t.Fatal("expected created task to have non-zero ID")
	}
	if resp.Task.Title != "Pay bills" {
		t.Fatalf("expected title %q, got %q", "Pay bills", resp.Task.Title)
	}
	if resp.Task.Description != "Electric + water" {
		t.Fatalf("expected description %q, got %q", "Electric + water", resp.Task.Description)
	}
	if resp.Task.State != models.Scheduled {
		t.Fatalf("expected state %q, got %q", models.Scheduled, resp.Task.State)
	}
}

func TestMakeCreateTask_InvalidDueDate(t *testing.T) {
	db := newTestTransportDB(t)
	router := newTestRouter(db)

	body := map[string]any{
		"title":       "Pay bills",
		"description": "Electric + water",
		"dueDate":     "03/07/2026",
	}

	w := performJSONRequest(router, http.MethodPost, "/api/tasks", body)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d; body=%s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestMakeUpdateTask_PartialUpdateSuccess(t *testing.T) {
	db := newTestTransportDB(t)
	router := newTestRouter(db)

	task := seedTask(t, db, "Original", time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC), models.Scheduled)

	newTitle := "Updated title"
	newState := "rescheduled"
	newDueDate := "2026-03-22"

	body := map[string]any{
		"title":   newTitle,
		"state":   newState,
		"dueDate": newDueDate,
	}

	w := performJSONRequest(router, http.MethodPatch, "/api/tasks/"+jsonUint(task.ID), body)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d; body=%s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp struct {
		Task models.Task `json:"task"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Task.Title != newTitle {
		t.Fatalf("expected title %q, got %q", newTitle, resp.Task.Title)
	}
	if resp.Task.State != models.Rescheduled {
		t.Fatalf("expected state %q, got %q", models.Rescheduled, resp.Task.State)
	}
	if !resp.Task.WasRescheduled {
		t.Fatal("expected WasRescheduled to be true")
	}
	if resp.Task.DueDate.Format("2006-01-02") != newDueDate {
		t.Fatalf("expected dueDate %q, got %q", newDueDate, resp.Task.DueDate.Format("2006-01-02"))
	}
}

func TestMakeUpdateTask_InvalidID(t *testing.T) {
	db := newTestTransportDB(t)
	router := newTestRouter(db)

	body := map[string]any{"title": "Updated"}

	w := performJSONRequest(router, http.MethodPatch, "/api/tasks/not-a-number", body)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d; body=%s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestMakeUpdateTask_NotFound(t *testing.T) {
	db := newTestTransportDB(t)
	router := newTestRouter(db)

	body := map[string]any{"title": "Updated"}

	w := performJSONRequest(router, http.MethodPatch, "/api/tasks/9999", body)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d; body=%s", http.StatusNotFound, w.Code, w.Body.String())
	}
}

func TestMakeDeleteTask_Success(t *testing.T) {
	db := newTestTransportDB(t)
	router := newTestRouter(db)

	task := seedTask(t, db, "Delete me", time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC), models.Scheduled)

	w := performJSONRequest(router, http.MethodDelete, "/api/tasks/"+jsonUint(task.ID), nil)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d; body=%s", http.StatusNoContent, w.Code, w.Body.String())
	}

	var found models.Task
	err := db.First(&found, task.ID).Error
	if err == nil {
		t.Fatal("expected task to be deleted, but it was still found")
	}
}

func TestMakeDeleteTask_InvalidID(t *testing.T) {
	db := newTestTransportDB(t)
	router := newTestRouter(db)

	w := performJSONRequest(router, http.MethodDelete, "/api/tasks/not-a-number", nil)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d; body=%s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func jsonUint(v uint) string {
	return fmt.Sprintf("%d", v)
}

func strconvFormatUint(v uint64) string {
	// small wrapper to avoid importing strconv all over tests
	return fmt.Sprintf("%d", v)
}
