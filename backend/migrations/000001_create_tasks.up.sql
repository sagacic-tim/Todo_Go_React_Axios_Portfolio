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

CREATE TABLE IF NOT EXISTS tasks (
  id            SERIAL        PRIMARY KEY,
  title         TEXT          NOT NULL,
  description   TEXT          NOT NULL,
  due_date      DATE          NOT NULL,
  state         tasks_state   NOT NULL  DEFAULT 'scheduled',
  rescheduled_at DATE,            -- when state='rescheduled'
  created_at    TIMESTAMPTZ    NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ    NOT NULL DEFAULT now()
);
