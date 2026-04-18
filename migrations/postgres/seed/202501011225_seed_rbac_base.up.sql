INSERT INTO roles (tenant_id, code, name, created_at, updated_at) VALUES
  ('default', 'admin', '管理员', NOW(), NOW()),
  ('default', 'user', '普通用户', NOW(), NOW())
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO role_permissions (tenant_id, role, permission, created_at, updated_at) VALUES
  ('default', 'admin', 'db:ping', NOW(), NOW()),
  ('default', 'admin', 'user:read', NOW(), NOW()),
  ('default', 'admin', 'user:create', NOW(), NOW()),
  ('default', 'admin', 'user:update', NOW(), NOW()),
  ('default', 'admin', 'user:delete', NOW(), NOW()),
  ('default', 'admin', 'user:export', NOW(), NOW()),
  ('default', 'admin', 'task:read', NOW(), NOW()),
  ('default', 'admin', 'task:create', NOW(), NOW()),
  ('default', 'admin', 'task:update', NOW(), NOW()),
  ('default', 'admin', 'task:delete', NOW(), NOW()),
  ('default', 'admin', 'task:toggle', NOW(), NOW()),
  ('default', 'admin', 'task:run', NOW(), NOW()),
  ('default', 'admin', 'menu:read', NOW(), NOW()),
  ('default', 'user', 'user:read', NOW(), NOW())
ON CONFLICT (tenant_id, role, permission) DO NOTHING;
