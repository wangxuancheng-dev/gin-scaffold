INSERT IGNORE INTO `roles` (`code`, `name`, `created_at`, `updated_at`) VALUES
  ('admin', '管理员', NOW(), NOW()),
  ('user', '普通用户', NOW(), NOW());

INSERT IGNORE INTO `role_permissions` (`role`, `permission`, `created_at`, `updated_at`) VALUES
  ('admin', 'db:ping', NOW(), NOW()),
  ('admin', 'user:read', NOW(), NOW()),
  ('admin', 'user:create', NOW(), NOW()),
  ('admin', 'user:update', NOW(), NOW()),
  ('admin', 'user:delete', NOW(), NOW()),
  ('admin', 'user:export', NOW(), NOW()),
  ('admin', 'task:read', NOW(), NOW()),
  ('admin', 'task:create', NOW(), NOW()),
  ('admin', 'task:update', NOW(), NOW()),
  ('admin', 'task:delete', NOW(), NOW()),
  ('admin', 'task:toggle', NOW(), NOW()),
  ('admin', 'task:run', NOW(), NOW()),
  ('admin', 'menu:read', NOW(), NOW()),
  ('user', 'user:read', NOW(), NOW());
