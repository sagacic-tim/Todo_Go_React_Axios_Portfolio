// backend/internal/transport/task_handler.go

package transport

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"todo_axios_api/backend/internal/models"
)

const (
	defaultTasksLimit = 500
	maxTasksLimit     = 5000
)

// GET /api/tasks?limit=500
// Returns JSON: { "tasks": [ /* array of TaskModel */ ] }
func makeGetTasks(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := defaultTasksLimit
		if s := c.Query("limit"); s != "" {
			if n, err := strconv.Atoi(s); err == nil {
				if n > 0 && n <= maxTasksLimit {
					limit = n
				}
			}
		}

		var tasks []models.Task

		// NOTE:
		// - Order gives stable results and benefits from an index on due_date.
		// - Limit prevents accidental "download the world" later.
		// If you later want to reduce payload further, switch to a summary struct + Select(...)
		// but that would change JSON unless you mirror the model fields.
		if err := db.
			Order("due_date ASC").
			Limit(limit).
			Find(&tasks).Error; err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"tasks": tasks})
	}
}

// bind JSON for POST /api/tasks
type createTaskInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	DueDate     string `json:"dueDate" binding:"required"` // expect "YYYY-MM-DD"
}

// POST /api/tasks
// Expects { title, description, dueDate }.
// Responds with { task: { … } } on 201.
func makeCreateTask(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in createTaskInput
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		d, err := time.Parse("2006-01-02", in.DueDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "dueDate must be YYYY-MM-DD"})
			return
		}

		t := models.Task{
			Title:       in.Title,
			Description: in.Description,
			DueDate:     d,
			State:       models.Scheduled,
		}

		if err := db.Create(&t).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"task": t})
	}
}

// PATCH /api/tasks/:id
//
// IMPORTANT CHANGE:
// - Supports partial updates (no longer requires dueDate on every PATCH).
func makeUpdateTask(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id64, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id64 == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		id := uint(id64)

		// Use pointers so we can distinguish “missing” vs “present but empty string”.
		var in struct {
			Title       *string `json:"title"`
			Description *string `json:"description"`
			DueDate     *string `json:"dueDate"`
			State       *string `json:"state"`
		}

		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// load existing
		var task models.Task
		if err := db.First(&task, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}

		// Apply updates only for fields provided.
		updates := map[string]any{}

		if in.Title != nil {
			updates["title"] = *in.Title
			task.Title = *in.Title
		}
		if in.Description != nil {
			updates["description"] = *in.Description
			task.Description = *in.Description
		}
		if in.DueDate != nil {
			d, err := time.Parse("2006-01-02", *in.DueDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "dueDate must be YYYY-MM-DD"})
				return
			}
			updates["due_date"] = d
			task.DueDate = d
		}
		if in.State != nil {
			// preserve history
			switch *in.State {
			case "rescheduled":
				updates["was_rescheduled"] = true
				task.WasRescheduled = true
			case "dismissed":
				updates["was_dismissed"] = true
				task.WasDismissed = true
			}
			updates["state"] = models.TaskState(*in.State)
			task.State = models.TaskState(*in.State)
		}

		if len(updates) == 0 {
			// Nothing to do, but return current value
			c.JSON(http.StatusOK, gin.H{"task": task})
			return
		}

		// Prefer Updates(map) over Save() to avoid rewriting the whole row unnecessarily.
		if err := db.Model(&models.Task{}).
			Where("id = ?", id).
			Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return the freshest version (optional; cheap)
		if err := db.First(&task, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"task": task})
	}
}

// DELETE /api/tasks/:id
func makeDeleteTask(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id64, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id64 == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		id := uint(id64)

		if err := db.Delete(&models.Task{}, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
