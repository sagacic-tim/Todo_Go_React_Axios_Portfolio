ALTER TABLE tasks
  DROP COLUMN IF EXISTS was_rescheduled,
  DROP COLUMN IF EXISTS was_dismissed;
