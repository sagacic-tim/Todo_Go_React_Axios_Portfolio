// backend/internal/models/tasks.go

package models

import "time"

// TaskState mirrors your Postgres ENUM
type TaskState string

const (
  Scheduled   TaskState = "scheduled"
  Rescheduled TaskState = "rescheduled"
  Completed   TaskState = "completed"
  Dismissed   TaskState = "dismissed"
)

// Task is your GORM model

type Task struct {
  ID             uint      `gorm:"primaryKey" json:"id"`
  Title          string    `gorm:"not null"    json:"title"`
  Description    string    `gorm:"not null"    json:"description"`
  DueDate        time.Time `gorm:"not null"    json:"dueDate"`
  State          TaskState `gorm:"type:tasks_state;not null;default:'scheduled'" json:"state"`
  WasRescheduled bool      `gorm:"not null;default:false" json:"wasRescheduled"`
  WasDismissed   bool      `gorm:"not null;default:false" json:"wasDismissed"`
}
