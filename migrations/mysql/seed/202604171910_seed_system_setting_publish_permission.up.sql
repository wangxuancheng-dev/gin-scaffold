INSERT IGNORE INTO `role_permissions` (`role`, `permission`, `created_at`, `updated_at`) VALUES
  ('admin', 'sys:config:publish', NOW(), NOW());
