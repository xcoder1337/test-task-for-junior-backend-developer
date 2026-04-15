CREATE TABLE IF NOT EXISTS tasks (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	due_date TIMESTAMPTZ NOT NULL DEFAULT '1970-01-01 00:00:00+00',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	
	-- Новые поля для периодичности
	recurrence_id BIGINT,
	is_recurrence_instance BOOLEAN DEFAULT FALSE,
	template_id TEXT
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks (due_date);
CREATE INDEX IF NOT EXISTS idx_tasks_template_id ON tasks (template_id);
CREATE INDEX IF NOT EXISTS idx_tasks_recurrence_id ON tasks (recurrence_id);

-- Таблица для правил повторения
CREATE TABLE IF NOT EXISTS recurrences (
	id BIGSERIAL PRIMARY KEY,
	task_template_id TEXT NOT NULL UNIQUE,
	type TEXT NOT NULL,
	value TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_recurrences_template_id ON recurrences (task_template_id);