INSERT INTO menus (id, tenant_id, name, path, perm_code, sort, parent_id, created_at, updated_at) VALUES
  (1, 'default', '用户管理', '/admin/users', 'user:read', 10, NULL, NOW(), NOW()),
  -- (2, 'default', '系统状态', '/admin/system', 'db:ping', 20, NULL, NOW(), NOW()),
  (3, 'default', '菜单管理', '/admin/menus', 'menu:read', 30, NULL, NOW(), NOW()),
  (4, 'default', '任务中心', '/admin/tasks', 'task:read', 40, NULL, NOW(), NOW()),
  (5, 'default', '审计日志', '/admin/audit-logs', 'audit:read', 50, NULL, NOW(), NOW()),
  (6, 'default', '系统参数', '/admin/system-settings', 'sys:config:read', 60, NULL, NOW(), NOW())
ON CONFLICT (tenant_id, path) DO NOTHING;

SELECT setval(pg_get_serial_sequence('menus', 'id'), (SELECT COALESCE(MAX(id), 1) FROM menus));
