// internal/repository/task_repo_gorm.go
package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"todo_axios_api/backend/internal/models"
)

type taskRepoGorm struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepoGorm{db: db}
}

func (r *taskRepoGorm) Create(ctx context.Context, t *models.Task) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *taskRepoGorm) FindAll(ctx context.Context) ([]models.Task, error) {
	var tasks []models.Task
	err := r.db.WithContext(ctx).
		Order("due_date ASC").
		Find(&tasks).Error
	return tasks, err
}

func (r *taskRepoGorm) FindByID(ctx context.Context, id uint) (*models.Task, error) {
	var task models.Task

	err := r.db.WithContext(ctx).
		First(&task, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}

	return &task, nil
}

func (r *taskRepoGorm) Update(ctx context.Context, t *models.Task) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *taskRepoGorm) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Task{}, id).Error
}
