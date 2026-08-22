// internal/repository/task_repo.go
package repository

import (
	"context"

	"todo_axios_api/backend/internal/models"
)

type TaskRepository interface {
	Create(ctx context.Context, t *models.Task) error
	FindAll(ctx context.Context) ([]models.Task, error)
	FindByID(ctx context.Context, id uint) (*models.Task, error)
	Update(ctx context.Context, t *models.Task) error
	Delete(ctx context.Context, id uint) error
}
