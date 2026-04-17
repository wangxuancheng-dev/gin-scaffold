INSERT IGNORE INTO `role_permissions` (`role`, `permission`, `created_at`, `updated_at`) VALUES
  ('admin', 'audit:read', NOW(), NOW());

INSERT IGNORE INTO `menus` (`id`, `name`, `path`, `perm_code`, `sort`, `created_at`, `updated_at`) VALUES
  (5, '审计日志', '/admin/audit-logs', 'audit:read', 50, NOW(), NOW());

INSERT IGNORE INTO `role_menus` (`role`, `menu_id`, `created_at`, `updated_at`) VALUES
  ('admin', 5, NOW(), NOW());
