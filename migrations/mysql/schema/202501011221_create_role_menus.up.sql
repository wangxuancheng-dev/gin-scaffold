CREATE TABLE IF NOT EXISTS `role_menus` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` VARCHAR(64) NOT NULL DEFAULT 'default',
  `role` VARCHAR(64) NOT NULL,
  `menu_id` BIGINT NOT NULL,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_tenant_menu` (`tenant_id`, `role`, `menu_id`),
  KEY `idx_role_menus_tenant` (`tenant_id`),
  KEY `idx_role_menus_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
