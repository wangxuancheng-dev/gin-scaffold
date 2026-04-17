ALTER TABLE `users`
ADD COLUMN `tenant_id` VARCHAR(64) NOT NULL DEFAULT 'default',
DROP INDEX `uk_users_username`,
ADD UNIQUE KEY `uk_users_tenant_username` (`tenant_id`, `username`),
ADD KEY `idx_users_tenant` (`tenant_id`);

ALTER TABLE `scheduled_tasks`
ADD COLUMN `tenant_id` VARCHAR(64) NOT NULL DEFAULT 'default',
DROP INDEX `uk_scheduled_tasks_name`,
ADD UNIQUE KEY `uk_scheduled_tasks_tenant_name` (`tenant_id`, `name`),
ADD KEY `idx_scheduled_tasks_tenant` (`tenant_id`);

ALTER TABLE `scheduled_task_logs`
ADD COLUMN `tenant_id` VARCHAR(64) NOT NULL DEFAULT 'default',
DROP INDEX `idx_scheduled_task_logs_task_id`,
ADD KEY `idx_scheduled_task_logs_tenant_task` (`tenant_id`, `task_id`);

ALTER TABLE `audit_logs`
ADD COLUMN `tenant_id` VARCHAR(64) NOT NULL DEFAULT 'default',
ADD KEY `idx_audit_tenant_created` (`tenant_id`, `created_at`);
