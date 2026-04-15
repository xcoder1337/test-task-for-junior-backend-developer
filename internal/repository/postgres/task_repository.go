package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type TaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

// ========== СТАРЫЕ МЕТОДЫ ==========

func (r *TaskRepository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, title, description, status, due_date, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query,
		task.Title, task.Description, task.Status, task.DueDate,
		task.CreatedAt, task.UpdatedAt)

	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, due_date, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}
		return nil, err
	}

	return found, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			due_date = $4,
			updated_at = $5
		WHERE id = $6
		RETURNING id, title, description, status, due_date, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query,
		task.Title, task.Description, task.Status, task.DueDate,
		task.UpdatedAt, task.ID)

	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}
		return nil, err
	}

	return updated, nil
}

func (r *TaskRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *TaskRepository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, due_date, created_at, updated_at
		FROM tasks
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// ========== НОВЫЕ МЕТОДЫ ДЛЯ ПЕРИОДИЧНОСТИ ==========

func (r *TaskRepository) CreateRecurrence(ctx context.Context, taskTemplateID, recurrenceType, value string) (int64, error) {
	const query = `
		INSERT INTO recurrences (task_template_id, type, value)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int64
	err := r.pool.QueryRow(ctx, query, taskTemplateID, recurrenceType, value).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *TaskRepository) CreateRecurrenceTask(ctx context.Context, task *taskdomain.Task, recurrenceID int64, templateID string) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, due_date, created_at, updated_at, recurrence_id, is_recurrence_instance, template_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, true, $8)
		RETURNING id, title, description, status, due_date, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query,
		task.Title, task.Description, task.Status, task.DueDate,
		task.CreatedAt, task.UpdatedAt,
		recurrenceID, templateID)

	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *TaskRepository) ExistsByTemplateAndDate(ctx context.Context, templateID string, dueDate interface{}) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1 FROM tasks 
			WHERE template_id = $1 AND due_date::date = $2::date
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, templateID, dueDate).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *TaskRepository) DeleteByTemplateID(ctx context.Context, templateID string) error {
	const query = `DELETE FROM tasks WHERE template_id = $1`

	_, err := r.pool.Exec(ctx, query, templateID)
	return err
}

// ========== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ==========

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task   taskdomain.Task
		status string
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	return &task, nil
}
