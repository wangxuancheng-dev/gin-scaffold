CREATE TABLE IF NOT EXISTS `scheduled_task_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` VARCHAR(64) NOT NULL DEFAULT 'default',
  `task_id` BIGINT NOT NULL,
  `status` VARCHAR(32) NOT NULL,
  `output` TEXT NULL,
  `error_message` TEXT NULL,
  `started_at` DATETIME(3) NOT NULL,
  `finished_at` DATETIME(3) NOT NULL,
  `duration_ms` BIGINT NOT NULL DEFAULT 0,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  KEY `idx_scheduled_task_logs_tenant_task` (`tenant_id`, `task_id`),
  KEY `idx_scheduled_task_logs_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
