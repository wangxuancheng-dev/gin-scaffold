CREATE TABLE IF NOT EXISTS `menus` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` VARCHAR(64) NOT NULL DEFAULT 'default',
  `name` VARCHAR(128) NOT NULL,
  `path` VARCHAR(255) NOT NULL,
  `perm_code` VARCHAR(128) NOT NULL,
  `sort` INT NOT NULL DEFAULT 0,
  `parent_id` BIGINT NULL DEFAULT NULL,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_menus_tenant_path` (`tenant_id`, `path`),
  KEY `idx_menus_tenant` (`tenant_id`),
  KEY `idx_menus_parent` (`tenant_id`, `parent_id`),
  KEY `idx_menus_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
