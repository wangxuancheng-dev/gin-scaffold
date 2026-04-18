DELETE FROM `role_permissions`
WHERE `tenant_id` = 'default' AND `role` = 'admin' AND `permission` IN ('menu:catalog', 'menu:create', 'menu:update', 'menu:delete');
