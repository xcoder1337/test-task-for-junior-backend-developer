package task

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)

	CreateRecurrence(ctx context.Context, taskTemplateID, recurrenceType, value string) (int64, error)
	CreateRecurrenceTask(ctx context.Context, task *taskdomain.Task, recurrenceID int64, templateID string) (*taskdomain.Task, error)
	ExistsByTemplateAndDate(ctx context.Context, templateID string, dueDate interface{}) (bool, error)
	DeleteByTemplateID(ctx context.Context, templateID string) error
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)

	CreateRecurringTasks(ctx context.Context, input CreateRecurringInput) ([]taskdomain.Task, error)
}

type CreateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	DueDate     time.Time
}

type UpdateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	DueDate     time.Time
}

type CreateRecurringInput struct {
	Title       string
	Description string
	DueDate     time.Time
	Recurrence  *taskdomain.RecurrenceConfig
}
