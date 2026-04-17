CREATE TABLE IF NOT EXISTS `system_settings` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `key` VARCHAR(128) NOT NULL,
  `value` TEXT NOT NULL,
  `value_type` VARCHAR(16) NOT NULL DEFAULT 'string',
  `group_name` VARCHAR(64) NOT NULL DEFAULT '',
  `remark` VARCHAR(255) NOT NULL DEFAULT '',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_system_settings_key` (`key`),
  KEY `idx_system_settings_group_name` (`group_name`),
  KEY `idx_system_settings_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
