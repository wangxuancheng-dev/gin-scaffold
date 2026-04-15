INSERT IGNORE INTO `menus` (`id`, `name`, `path`, `perm_code`, `sort`, `created_at`, `updated_at`) VALUES
  (1, '用户管理', '/admin/users', 'user:read', 10, NOW(), NOW()),
  (2, '系统状态', '/admin/system', 'db:ping', 20, NOW(), NOW()),
  (3, '菜单管理', '/admin/menus', 'menu:read', 30, NOW(), NOW()),
  (4, '任务中心', '/admin/tasks', 'task:read', 40, NOW(), NOW());

INSERT IGNORE INTO `role_menus` (`role`, `menu_id`, `created_at`, `updated_at`) VALUES
  ('admin', 1, NOW(), NOW()),
  ('admin', 2, NOW(), NOW()),
  ('admin', 3, NOW(), NOW()),
  ('admin', 4, NOW(), NOW());
