INSERT INTO role_permissions (tenant_id, role, permission, created_at, updated_at) VALUES
  ('default', 'admin', 'sys:config:read', NOW(), NOW()),
  ('default', 'admin', 'sys:config:write', NOW(), NOW())
ON CONFLICT (tenant_id, role, permission) DO NOTHING;

INSERT INTO menus (id, tenant_id, name, path, perm_code, sort, parent_id, created_at, updated_at) VALUES
  (6, 'default', '系统参数', '/admin/system-settings', 'sys:config:read', 60, NULL, NOW(), NOW())
ON CONFLICT (tenant_id, path) DO NOTHING;

INSERT INTO role_menus (tenant_id, role, menu_id, created_at, updated_at) VALUES
  ('default', 'admin', 6, NOW(), NOW())
ON CONFLICT (tenant_id, role, menu_id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('menus', 'id'), (SELECT COALESCE(MAX(id), 1) FROM menus));
