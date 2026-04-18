INSERT INTO role_permissions (tenant_id, role, permission, created_at, updated_at) VALUES
  ('default', 'admin', 'announcement:read', NOW(), NOW()),
  ('default', 'admin', 'announcement:write', NOW(), NOW())
ON CONFLICT (tenant_id, role, permission) DO NOTHING;
