package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

// ========== СТАРЫЕ МЕТОДЫ (НЕ ТРОГАЕМ) ==========

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	// ЛОГИРОВАНИЕ
	fmt.Printf("DEBUG: Create called with DueDate = %v\n", input.DueDate)

	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	fmt.Printf("DEBUG: After validation DueDate = %v\n", normalized.DueDate)

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		DueDate:     normalized.DueDate,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	fmt.Printf("DEBUG: Before repo.Create DueDate = %v\n", model.DueDate)

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	fmt.Printf("DEBUG: After repo.Create DueDate = %v\n", created.DueDate)

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		DueDate:     normalized.DueDate,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

// ========== НОВЫЙ МЕТОД ДЛЯ ПЕРИОДИЧЕСКИХ ЗАДАЧ ==========

// CreateRecurringTasks создает периодические задачи на основе recurrence конфига
func (s *Service) CreateRecurringTasks(ctx context.Context, input CreateRecurringInput) ([]taskdomain.Task, error) {
	// Валидация базовых полей
	if strings.TrimSpace(input.Title) == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.DueDate.IsZero() {
		return nil, fmt.Errorf("%w: due_date is required", ErrInvalidInput)
	}

	if input.Recurrence == nil {
		return nil, fmt.Errorf("%w: recurrence config is required", ErrInvalidInput)
	}

	// Генерируем даты
	generator := NewRecurrenceGenerator()
	dates, err := generator.GenerateDates(*input.Recurrence, input.DueDate)
	if err != nil {
		return nil, fmt.Errorf("failed to generate dates: %w", err)
	}

	if len(dates) == 0 {
		return nil, fmt.Errorf("%w: no dates generated for this recurrence rule", ErrInvalidInput)
	}

	// Создаем шаблон (rule)
	templateID := uuid.New().String()
	valueJSON, _ := json.Marshal(input.Recurrence)

	recurrenceID, err := s.repo.CreateRecurrence(ctx, templateID, input.Recurrence.Type, string(valueJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create recurrence rule: %w", err)
	}

	// Создаем экземпляры задач
	var createdTasks []taskdomain.Task
	now := s.now()

	for _, date := range dates {
		// Проверяем дубликат
		exists, _ := s.repo.ExistsByTemplateAndDate(ctx, templateID, date)
		if exists {
			continue
		}

		task := &taskdomain.Task{
			Title:       input.Title,
			Description: input.Description,
			Status:      taskdomain.StatusNew,
			DueDate:     date,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		created, err := s.repo.CreateRecurrenceTask(ctx, task, recurrenceID, templateID)
		if err != nil {
			return nil, fmt.Errorf("failed to create task for date %s: %w", date.Format("2006-01-02"), err)
		}

		createdTasks = append(createdTasks, *created)
	}

	if len(createdTasks) == 0 {
		return nil, fmt.Errorf("%w: all tasks already exist (duplicates)", ErrInvalidInput)
	}

	return createdTasks, nil
}

// RecurrenceGenerator интерфейс (реальная реализация в services пакете)
type RecurrenceGeneratorInterface interface {
	GenerateDates(config taskdomain.RecurrenceConfig, startDate time.Time) ([]time.Time, error)
}

// Временная реализация (потом заменим на импорт из services)
type RecurrenceGenerator struct{}

func NewRecurrenceGenerator() *RecurrenceGenerator {
	return &RecurrenceGenerator{}
}

func (g *RecurrenceGenerator) GenerateDates(config taskdomain.RecurrenceConfig, startDate time.Time) ([]time.Time, error) {
	lookahead := config.LookaheadDays
	if lookahead <= 0 {
		lookahead = 90
	}
	endDate := startDate.AddDate(0, 0, lookahead)

	switch config.Type {
	case "interval":
		return g.generateInterval(config.EveryNDays, startDate, endDate), nil
	case "monthly":
		return g.generateMonthly(config.DayOfMonth, startDate, endDate), nil
	case "fixed_dates":
		return g.generateFixedDates(config.Dates, startDate, endDate)
	case "parity":
		return g.generateParity(config.Parity, startDate, endDate), nil
	default:
		return nil, fmt.Errorf("unknown recurrence type: %s", config.Type)
	}
}

func (g *RecurrenceGenerator) generateInterval(everyNDays int, start, end time.Time) []time.Time {
	if everyNDays <= 0 {
		everyNDays = 1
	}
	var dates []time.Time
	current := start
	for !current.After(end) {
		dates = append(dates, current)
		current = current.AddDate(0, 0, everyNDays)
	}
	return dates
}

func (g *RecurrenceGenerator) generateMonthly(dayOfMonth int, start, end time.Time) []time.Time {
	if dayOfMonth < 1 {
		dayOfMonth = 1
	}
	if dayOfMonth > 30 {
		dayOfMonth = 30
	}
	var dates []time.Time
	current := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	for !current.After(end) {
		targetDay := dayOfMonth
		lastDayOfMonth := time.Date(current.Year(), current.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		if targetDay > lastDayOfMonth {
			targetDay = lastDayOfMonth
		}
		candidate := time.Date(current.Year(), current.Month(), targetDay, 0, 0, 0, 0, time.UTC)
		if !candidate.Before(start) && !candidate.After(end) {
			dates = append(dates, candidate)
		}
		current = current.AddDate(0, 1, 0)
	}
	return dates
}

func (g *RecurrenceGenerator) generateFixedDates(datesStr []string, start, end time.Time) ([]time.Time, error) {
	var dates []time.Time
	for _, ds := range datesStr {
		t, err := time.Parse("2006-01-02", ds)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %s", ds)
		}
		t = t.UTC()
		if !t.Before(start) && !t.After(end) {
			dates = append(dates, t)
		}
	}
	return dates, nil
}

func (g *RecurrenceGenerator) generateParity(parity string, start, end time.Time) []time.Time {
	var dates []time.Time
	current := start
	for !current.After(end) {
		day := current.Day()
		if (parity == "even" && day%2 == 0) || (parity == "odd" && day%2 == 1) {
			dates = append(dates, current)
		}
		current = current.AddDate(0, 0, 1)
	}
	return dates
}

// ========== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ==========

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	// DueDate не нужно валидировать, оно может быть zero value
	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}
