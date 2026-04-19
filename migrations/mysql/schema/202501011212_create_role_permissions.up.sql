CREATE TABLE IF NOT EXISTS `role_permissions` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` VARCHAR(64) NOT NULL DEFAULT 'default',
  `role` VARCHAR(64) NOT NULL,
  `permission` VARCHAR(128) NOT NULL,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_tenant_permission` (`tenant_id`, `role`, `permission`),
  KEY `idx_role_permissions_tenant` (`tenant_id`),
  KEY `idx_role_permissions_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
