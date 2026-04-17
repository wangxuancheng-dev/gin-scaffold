CREATE TABLE IF NOT EXISTS `audit_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `request_id` VARCHAR(64) NULL,
  `user_id` BIGINT NOT NULL DEFAULT 0,
  `role` VARCHAR(32) NULL,
  `actor_type` VARCHAR(16) NOT NULL DEFAULT 'anonymous',
  `action` VARCHAR(16) NOT NULL,
  `path` VARCHAR(512) NOT NULL,
  `query` VARCHAR(1024) NULL,
  `status` INT NOT NULL,
  `latency_ms` INT NOT NULL DEFAULT 0,
  `client_ip` VARCHAR(64) NULL,
  `created_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  KEY `idx_audit_request` (`request_id`),
  KEY `idx_audit_created` (`created_at`),
  KEY `idx_audit_user_created` (`user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
