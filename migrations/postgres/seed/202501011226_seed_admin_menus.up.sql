INSERT INTO menus (id, tenant_id, name, path, perm_code, sort, parent_id, created_at, updated_at) VALUES
  (1, 'default', '用户管理', '/admin/users', 'user:read', 10, NULL, NOW(), NOW()),
  (2, 'default', '系统状态', '/admin/system', 'db:ping', 20, NULL, NOW(), NOW()),
  (3, 'default', '菜单管理', '/admin/menus', 'menu:read', 30, NULL, NOW(), NOW()),
  (4, 'default', '任务中心', '/admin/tasks', 'task:read', 40, NULL, NOW(), NOW())
ON CONFLICT (tenant_id, path) DO NOTHING;

INSERT INTO role_menus (tenant_id, role, menu_id, created_at, updated_at) VALUES
  ('default', 'admin', 1, NOW(), NOW()),
  ('default', 'admin', 2, NOW(), NOW()),
  ('default', 'admin', 3, NOW(), NOW()),
  ('default', 'admin', 4, NOW(), NOW())
ON CONFLICT (tenant_id, role, menu_id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('menus', 'id'), (SELECT COALESCE(MAX(id), 1) FROM menus));
