CREATE TABLE IF NOT EXISTS `scheduled_tasks` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL,
  `spec` VARCHAR(64) NOT NULL,
  `command` VARCHAR(1024) NOT NULL,
  `timeout_sec` INT NOT NULL DEFAULT 30,
  `enabled` TINYINT(1) NOT NULL DEFAULT 1,
  `last_run_at` DATETIME(3) NULL,
  `last_status` VARCHAR(32) NULL,
  `last_message` VARCHAR(255) NULL,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_scheduled_tasks_name` (`name`),
  KEY `idx_scheduled_tasks_enabled` (`enabled`),
  KEY `idx_scheduled_tasks_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `scheduled_task_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
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
  KEY `idx_scheduled_task_logs_task_id` (`task_id`),
  KEY `idx_scheduled_task_logs_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
