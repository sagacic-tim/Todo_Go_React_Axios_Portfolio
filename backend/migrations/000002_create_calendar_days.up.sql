
-- 000002_create_calendar_days.up.sql
-- Creates calendar_days for mapping a date to a task

CREATE TABLE IF NOT EXISTS calendar_days (
  id       SERIAL PRIMARY KEY,
  year     INTEGER NOT NULL,
  month    INTEGER NOT NULL,
  day      INTEGER NOT NULL,
  task_id  INTEGER,
  CONSTRAINT fk_task
    FOREIGN KEY(task_id)
      REFERENCES tasks(id)
      ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
