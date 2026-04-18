ALTER TABLE scheduled_tasks
  DROP COLUMN IF EXISTS concurrency_policy;
