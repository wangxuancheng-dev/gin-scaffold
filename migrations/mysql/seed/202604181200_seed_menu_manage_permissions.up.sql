INSERT IGNORE INTO `role_permissions` (`tenant_id`, `role`, `permission`, `created_at`, `updated_at`) VALUES
  ('default', 'admin', 'menu:catalog', NOW(), NOW()),
  ('default', 'admin', 'menu:create', NOW(), NOW()),
  ('default', 'admin', 'menu:update', NOW(), NOW()),
  ('default', 'admin', 'menu:delete', NOW(), NOW());
