-- 1) make sure the enum type exists
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
      FROM pg_type
     WHERE typname = 'tasks_state'
  ) THEN
    CREATE TYPE tasks_state AS ENUM (
      'scheduled',
      'rescheduled',
      'completed',
      'dismissed'
    );
  END IF;
END
$$;

-- 2) add the state column (with default 'scheduled')
ALTER TABLE tasks
  ADD COLUMN IF NOT EXISTS state tasks_state NOT NULL DEFAULT 'scheduled',
  ADD COLUMN IF NOT EXISTS rescheduled_at DATE,
  ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
