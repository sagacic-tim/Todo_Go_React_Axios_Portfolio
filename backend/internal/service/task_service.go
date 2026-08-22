// backe nd/internal/service/task_service.go

package service

import (
	"context"
	"time"

	"todo_axios_api/backend/internal/models"
	"todo_axios_api/backend/internal/repository"
)

type TaskService interface {
	CreateTask(ctx context.Context, title, desc string, dueDate time.Time) (*models.Task, error)
	ListTasks(ctx context.Context) ([]models.Task, error)
}

type taskService struct {
	repo repository.TaskRepository
}

func NewTaskService(r repository.TaskRepository) TaskService {
	return &taskService{repo: r}
}

func (s *taskService) CreateTask(ctx context.Context, title, desc string, dueDate time.Time) (*models.Task, error) {
	t := &models.Task{
		Title:       title,
		Description: desc,
		DueDate:     dueDate,
		State:       "scheduled",
	}

	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *taskService) ListTasks(ctx context.Context) ([]models.Task, error) {
	return s.repo.FindAll(ctx)
}
