DELETE FROM role_permissions
WHERE tenant_id = 'default' AND (
  (role = 'admin' AND permission IN (
    'db:ping', 'user:read', 'user:create', 'user:update', 'user:delete', 'user:export',
    'task:read', 'task:create', 'task:update', 'task:delete', 'task:toggle', 'task:run', 'menu:read'
  ))
  OR (role = 'user' AND permission = 'user:read')
);

DELETE FROM roles
WHERE tenant_id = 'default' AND code IN ('admin', 'user');
