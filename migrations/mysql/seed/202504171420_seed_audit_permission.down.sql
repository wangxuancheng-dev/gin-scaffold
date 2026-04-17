DELETE FROM `role_menus` WHERE `role` = 'admin' AND `menu_id` = 5;
DELETE FROM `menus` WHERE `id` = 5 AND `perm_code` = 'audit:read';
DELETE FROM `role_permissions` WHERE `role` = 'admin' AND `permission` = 'audit:read';
