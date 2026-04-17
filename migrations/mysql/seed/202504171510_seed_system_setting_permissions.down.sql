DELETE FROM `role_menus` WHERE `role` = 'admin' AND `menu_id` = 6;
DELETE FROM `menus` WHERE `id` = 6 AND `perm_code` = 'sys:config:read';
DELETE FROM `role_permissions` WHERE `role` = 'admin' AND `permission` IN ('sys:config:read', 'sys:config:write');
