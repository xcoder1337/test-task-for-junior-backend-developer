package task

import (
	"time"
)

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	DueDate     time.Time `json:"due_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Recurrence fields
	RecurrenceID         *int64  `json:"recurrence_id,omitempty"`
	IsRecurrenceInstance bool    `json:"is_recurrence_instance"`
	TemplateID           *string `json:"template_id,omitempty"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

// RecurrenceRule хранит правило повторения
type RecurrenceRule struct {
	ID             int64     `json:"id"`
	TaskTemplateID string    `json:"task_template_id"`
	Type           string    `json:"type"`
	Value          string    `json:"value"`
	CreatedAt      time.Time `json:"created_at"`
}

// RecurrenceConfig для парсинга JSON из запроса
type RecurrenceConfig struct {
	Type          string   `json:"type"`
	EveryNDays    int      `json:"every_n_days,omitempty"`
	DayOfMonth    int      `json:"day_of_month,omitempty"`
	Dates         []string `json:"dates,omitempty"`
	Parity        string   `json:"parity,omitempty"`
	LookaheadDays int      `json:"lookahead_days,omitempty"`
}
