// backend/internal/service/task_service_test.go

package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"todo_axios_api/backend/internal/models"
)

// fakeTaskRepository is a test double for repository.TaskRepository.
type fakeTaskRepository struct {
	createFn   func(ctx context.Context, t *models.Task) error
	findAllFn  func(ctx context.Context) ([]models.Task, error)
	findByIDFn func(ctx context.Context, id uint) (*models.Task, error)
	updateFn   func(ctx context.Context, t *models.Task) error
	deleteFn   func(ctx context.Context, id uint) error
}

func (f *fakeTaskRepository) Create(ctx context.Context, t *models.Task) error {
	if f.createFn != nil {
		return f.createFn(ctx, t)
	}
	return nil
}

func (f *fakeTaskRepository) FindAll(ctx context.Context) ([]models.Task, error) {
	if f.findAllFn != nil {
		return f.findAllFn(ctx)
	}
	return nil, nil
}

func (f *fakeTaskRepository) FindByID(ctx context.Context, id uint) (*models.Task, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (f *fakeTaskRepository) Update(ctx context.Context, t *models.Task) error {
	if f.updateFn != nil {
		return f.updateFn(ctx, t)
	}
	return nil
}

func (f *fakeTaskRepository) Delete(ctx context.Context, id uint) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, id)
	}
	return nil
}

func TestNewTaskService(t *testing.T) {
	repo := &fakeTaskRepository{}
	svc := NewTaskService(repo)

	if svc == nil {
		t.Fatal("expected NewTaskService to return a non-nil service")
	}
}

func TestTaskService_CreateTask_Success(t *testing.T) {
	ctx := context.Background()
	dueDate := time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC)

	var created *models.Task

	repo := &fakeTaskRepository{
		createFn: func(ctx context.Context, t *models.Task) error {
			created = t
			return nil
		},
	}

	svc := NewTaskService(repo)

	got, err := svc.CreateTask(ctx, "Pay bills", "Electric + water", dueDate)
	if err != nil {
		t.Fatalf("CreateTask returned unexpected error: %v", err)
	}

	if got == nil {
		t.Fatal("expected non-nil task")
	}

	if created == nil {
		t.Fatal("expected repository Create to be called")
	}

	if got != created {
		t.Fatal("expected returned task pointer to match created task pointer")
	}

	if got.Title != "Pay bills" {
		t.Fatalf("expected Title %q, got %q", "Pay bills", got.Title)
	}

	if got.Description != "Electric + water" {
		t.Fatalf("expected Description %q, got %q", "Electric + water", got.Description)
	}

	if !got.DueDate.Equal(dueDate) {
		t.Fatalf("expected DueDate %v, got %v", dueDate, got.DueDate)
	}

	if got.State != models.Scheduled {
		t.Fatalf("expected State %q, got %q", models.Scheduled, got.State)
	}

	if got.WasRescheduled {
		t.Fatal("expected WasRescheduled to default to false")
	}

	if got.WasDismissed {
		t.Fatal("expected WasDismissed to default to false")
	}
}

func TestTaskService_CreateTask_PropagatesRepositoryError(t *testing.T) {
	ctx := context.Background()
	dueDate := time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC)
	expectedErr := errors.New("repository create failed")

	repo := &fakeTaskRepository{
		createFn: func(ctx context.Context, t *models.Task) error {
			return expectedErr
		},
	}

	svc := NewTaskService(repo)

	got, err := svc.CreateTask(ctx, "Pay bills", "Electric + water", dueDate)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Fatalf("expected nil task on error, got %#v", got)
	}
}

func TestTaskService_ListTasks_Success(t *testing.T) {
	ctx := context.Background()

	expected := []models.Task{
		{
			ID:             1,
			Title:          "Task 1",
			Description:    "First task",
			DueDate:        time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
			State:          models.Scheduled,
			WasRescheduled: false,
			WasDismissed:   false,
		},
		{
			ID:             2,
			Title:          "Task 2",
			Description:    "Second task",
			DueDate:        time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC),
			State:          models.Completed,
			WasRescheduled: false,
			WasDismissed:   false,
		},
	}

	repo := &fakeTaskRepository{
		findAllFn: func(ctx context.Context) ([]models.Task, error) {
			return expected, nil
		},
	}

	svc := NewTaskService(repo)

	got, err := svc.ListTasks(ctx)
	if err != nil {
		t.Fatalf("ListTasks returned unexpected error: %v", err)
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %#v, got %#v", expected, got)
	}
}

func TestTaskService_ListTasks_PropagatesRepositoryError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("repository FindAll failed")

	repo := &fakeTaskRepository{
		findAllFn: func(ctx context.Context) ([]models.Task, error) {
			return nil, expectedErr
		},
	}

	svc := NewTaskService(repo)

	got, err := svc.ListTasks(ctx)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Fatalf("expected nil tasks on error, got %#v", got)
	}
}
