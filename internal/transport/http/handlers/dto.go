package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	DueDate     string            `json:"due_date"` // НОВОЕ ПОЛЕ
}

type taskDTO struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	DueDate     time.Time         `json:"due_date"` // НОВОЕ ПОЛЕ
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// НОВЫЙ DTO для периодических задач
type recurringTaskRequest struct {
	Title       string                       `json:"title"`
	Description string                       `json:"description"`
	DueDate     string                       `json:"due_date"`
	Recurrence  *taskdomain.RecurrenceConfig `json:"recurrence"`
}

// НОВЫЙ DTO для ответа (список задач)
type recurringTaskResponse struct {
	Tasks []taskDTO `json:"tasks"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		DueDate:     task.DueDate,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

// НОВАЯ функция для преобразования списка задач
func newTasksDTO(tasks []taskdomain.Task) []taskDTO {
	result := make([]taskDTO, len(tasks))
	for i, task := range tasks {
		result[i] = newTaskDTO(&task)
	}
	return result
}
