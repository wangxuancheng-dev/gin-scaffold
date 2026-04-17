CREATE TABLE IF NOT EXISTS `outbox_events` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` VARCHAR(64) NOT NULL DEFAULT 'default',
  `topic` VARCHAR(128) NOT NULL,
  `payload` LONGTEXT NOT NULL,
  `status` VARCHAR(16) NOT NULL DEFAULT 'pending',
  `attempts` INT NOT NULL DEFAULT 0,
  `max_attempts` INT NOT NULL DEFAULT 10,
  `next_run_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `last_error` VARCHAR(512) NOT NULL DEFAULT '',
  `published_at` DATETIME(3) NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_outbox_tenant_status_next` (`tenant_id`, `status`, `next_run_at`),
  KEY `idx_outbox_topic_created` (`topic`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
