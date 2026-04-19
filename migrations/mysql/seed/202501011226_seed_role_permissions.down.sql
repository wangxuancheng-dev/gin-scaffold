DELETE FROM `role_permissions`
WHERE `tenant_id` = 'default'
  AND (
    (`role` = 'admin' AND `permission` IN (
      'db:ping', 'user:read', 'user:create', 'user:update', 'user:delete', 'user:export',
      'task:read', 'task:create', 'task:update', 'task:delete', 'task:toggle', 'task:run', 'menu:read',
      'audit:read', 'audit:export',
      'sys:config:read', 'sys:config:write', 'sys:config:rollback', 'sys:config:publish',
      'announcement:read', 'announcement:write',
      'menu:catalog', 'menu:create', 'menu:update', 'menu:delete'
    ))
    OR (`role` = 'user' AND `permission` = 'user:read')
  );
