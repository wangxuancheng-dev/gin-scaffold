INSERT INTO role_permissions (tenant_id, role, permission, created_at, updated_at) VALUES
  ('default', 'admin', 'sys:config:rollback', NOW(), NOW())
ON CONFLICT (tenant_id, role, permission) DO NOTHING;
