DELETE FROM `role_permissions`
WHERE (`role` = 'admin' AND `permission` IN ('db:ping', 'user:read', 'user:create', 'user:update', 'user:delete', 'user:export', 'menu:read'))
   OR (`role` = 'user' AND `permission` = 'user:read');

DELETE FROM `roles`
WHERE `code` IN ('admin', 'user');
