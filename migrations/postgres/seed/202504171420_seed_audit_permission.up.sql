INSERT INTO role_permissions (tenant_id, role, permission, created_at, updated_at) VALUES
  ('default', 'admin', 'audit:read', NOW(), NOW())
ON CONFLICT (tenant_id, role, permission) DO NOTHING;

INSERT INTO menus (id, tenant_id, name, path, perm_code, sort, parent_id, created_at, updated_at) VALUES
  (5, 'default', '审计日志', '/admin/audit-logs', 'audit:read', 50, NULL, NOW(), NOW())
ON CONFLICT (tenant_id, path) DO NOTHING;

INSERT INTO role_menus (tenant_id, role, menu_id, created_at, updated_at) VALUES
  ('default', 'admin', 5, NOW(), NOW())
ON CONFLICT (tenant_id, role, menu_id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('menus', 'id'), (SELECT COALESCE(MAX(id), 1) FROM menus));
