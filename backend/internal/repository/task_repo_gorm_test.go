// backend/internal/repository/task_repo_gorm_test.go

package repository

import (
	"context"
	"errors"
	"testing"
	"time"
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"todo_axios_api/backend/internal/models"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test sqlite database: %v", err)
	}

	if err := db.AutoMigrate(&models.Task{}); err != nil {
		t.Fatalf("failed to migrate test schema: %v", err)
	}

	return db
}

func seedTask(t *testing.T, db *gorm.DB, title string, dueDate time.Time, state models.TaskState) models.Task {
	t.Helper()

	task := models.Task{
		Title:          title,
		Description:    title + " description",
		DueDate:        dueDate,
		State:          state,
		WasRescheduled: false,
		WasDismissed:   false,
	}

	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}

	return task
}

func TestTaskRepoGorm_Create(t *testing.T) {
	db := newTestDB(t)
	repo := NewTaskRepository(db)

	ctx := context.Background()
	dueDate := time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC)

	task := &models.Task{
		Title:       "Pay bills",
		Description: "Electric and water",
		DueDate:     dueDate,
		State:       models.Scheduled,
	}

	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}

	if task.ID == 0 {
		t.Fatal("expected task ID to be set after Create")
	}

	var got models.Task
	if err := db.First(&got, task.ID).Error; err != nil {
		t.Fatalf("failed to fetch created task: %v", err)
	}

	if got.Title != task.Title {
		t.Fatalf("expected Title %q, got %q", task.Title, got.Title)
	}
	if got.Description != task.Description {
		t.Fatalf("expected Description %q, got %q", task.Description, got.Description)
	}
	if !got.DueDate.Equal(task.DueDate) {
		t.Fatalf("expected DueDate %v, got %v", task.DueDate, got.DueDate)
	}
	if got.State != task.State {
		t.Fatalf("expected State %q, got %q", task.State, got.State)
	}
}

func TestTaskRepoGorm_FindAll_OrdersByDueDateAsc(t *testing.T) {
	db := newTestDB(t)
	repo := NewTaskRepository(db)

	later := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)
	earlier := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)

	lateTask := seedTask(t, db, "Later task", later, models.Scheduled)
	earlyTask := seedTask(t, db, "Earlier task", earlier, models.Scheduled)

	ctx := context.Background()

	tasks, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll returned unexpected error: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}

	if tasks[0].ID != earlyTask.ID {
		t.Fatalf("expected first task ID %d, got %d", earlyTask.ID, tasks[0].ID)
	}
	if tasks[1].ID != lateTask.ID {
		t.Fatalf("expected second task ID %d, got %d", lateTask.ID, tasks[1].ID)
	}
}

func TestTaskRepoGorm_FindByID_Success(t *testing.T) {
	db := newTestDB(t)
	repo := NewTaskRepository(db)

	expected := seedTask(
		t,
		db,
		"Find me",
		time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC),
		models.Scheduled,
	)

	ctx := context.Background()

	got, err := repo.FindByID(ctx, expected.ID)
	if err != nil {
		t.Fatalf("FindByID returned unexpected error: %v", err)
	}

	if got == nil {
		t.Fatal("expected non-nil task")
	}

	if got.ID != expected.ID {
		t.Fatalf("expected ID %d, got %d", expected.ID, got.ID)
	}
	if got.Title != expected.Title {
		t.Fatalf("expected Title %q, got %q", expected.Title, got.Title)
	}
}

func TestTaskRepoGorm_FindByID_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewTaskRepository(db)

	ctx := context.Background()

	got, err := repo.FindByID(ctx, 9999)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil task, got %#v", got)
	}
}

func TestTaskRepoGorm_Update(t *testing.T) {
	db := newTestDB(t)
	repo := NewTaskRepository(db)

	task := seedTask(
		t,
		db,
		"Original title",
		time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		models.Scheduled,
	)

	ctx := context.Background()

	task.Title = "Updated title"
	task.Description = "Updated description"
	task.State = models.Completed
	task.WasDismissed = true

	if err := repo.Update(ctx, &task); err != nil {
		t.Fatalf("Update returned unexpected error: %v", err)
	}

	var got models.Task
	if err := db.First(&got, task.ID).Error; err != nil {
		t.Fatalf("failed to fetch updated task: %v", err)
	}

	if got.Title != "Updated title" {
		t.Fatalf("expected updated title, got %q", got.Title)
	}
	if got.Description != "Updated description" {
		t.Fatalf("expected updated description, got %q", got.Description)
	}
	if got.State != models.Completed {
		t.Fatalf("expected updated state %q, got %q", models.Completed, got.State)
	}
	if !got.WasDismissed {
		t.Fatal("expected WasDismissed to be true")
	}
}

func TestTaskRepoGorm_Delete(t *testing.T) {
	db := newTestDB(t)
	repo := NewTaskRepository(db)

	task := seedTask(
		t,
		db,
		"Delete me",
		time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC),
		models.Scheduled,
	)

	ctx := context.Background()

	if err := repo.Delete(ctx, task.ID); err != nil {
		t.Fatalf("Delete returned unexpected error: %v", err)
	}

	var got models.Task
	err := db.First(&got, task.ID).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound after delete, got %v", err)
	}
}
