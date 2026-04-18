ALTER TABLE scheduled_tasks
  ADD COLUMN IF NOT EXISTS concurrency_policy VARCHAR(16) NOT NULL DEFAULT 'forbid';
