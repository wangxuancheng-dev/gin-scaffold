ALTER TABLE `audit_logs`
DROP INDEX `idx_audit_tenant_created`,
DROP COLUMN `tenant_id`;

ALTER TABLE `scheduled_task_logs`
DROP INDEX `idx_scheduled_task_logs_tenant_task`,
ADD KEY `idx_scheduled_task_logs_task_id` (`task_id`),
DROP COLUMN `tenant_id`;

ALTER TABLE `scheduled_tasks`
DROP INDEX `uk_scheduled_tasks_tenant_name`,
DROP INDEX `idx_scheduled_tasks_tenant`,
ADD UNIQUE KEY `uk_scheduled_tasks_name` (`name`),
DROP COLUMN `tenant_id`;

ALTER TABLE `users`
DROP INDEX `uk_users_tenant_username`,
DROP INDEX `idx_users_tenant`,
ADD UNIQUE KEY `uk_users_username` (`username`),
DROP COLUMN `tenant_id`;
