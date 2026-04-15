CREATE TABLE IF NOT EXISTS `menus` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL,
  `path` VARCHAR(255) NOT NULL,
  `perm_code` VARCHAR(128) NOT NULL,
  `sort` INT NOT NULL DEFAULT 0,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_menus_path` (`path`),
  KEY `idx_menus_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `role_menus` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `role` VARCHAR(64) NOT NULL,
  `menu_id` BIGINT NOT NULL,
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_menu` (`role`, `menu_id`),
  KEY `idx_role_menus_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
