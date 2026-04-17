ALTER TABLE `scheduled_tasks`
ADD COLUMN `concurrency_policy` VARCHAR(16) NOT NULL DEFAULT 'forbid';
