INSERT IGNORE INTO role_permissions (role, permission, created_at, updated_at) VALUES
  ('admin', 'announcement:read', NOW(), NOW()),
  ('admin', 'announcement:write', NOW(), NOW());
