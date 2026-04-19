CREATE TABLE IF NOT EXISTS `user_roles` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` VARCHAR(64) NOT NULL DEFAULT 'default',
  `user_id` BIGINT NOT NULL,
  `role` VARCHAR(64) NOT NULL,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_tenant_role` (`tenant_id`, `user_id`, `role`),
  KEY `idx_user_roles_tenant` (`tenant_id`),
  KEY `idx_user_roles_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
