INSERT INTO roles (tenant_id, code, name, created_at, updated_at) VALUES
  ('default', 'admin', '管理员', NOW(), NOW()),
  ('default', 'user', '普通用户', NOW(), NOW())
ON CONFLICT (tenant_id, code) DO NOTHING;
