INSERT IGNORE INTO `role_permissions` (`role`, `permission`, `created_at`, `updated_at`) VALUES
  ('admin', 'sys:config:read', NOW(), NOW()),
  ('admin', 'sys:config:write', NOW(), NOW());

INSERT IGNORE INTO `menus` (`id`, `name`, `path`, `perm_code`, `sort`, `created_at`, `updated_at`) VALUES
  (6, '系统参数', '/admin/system-settings', 'sys:config:read', 60, NOW(), NOW());

INSERT IGNORE INTO `role_menus` (`role`, `menu_id`, `created_at`, `updated_at`) VALUES
  ('admin', 6, NOW(), NOW());
