DELETE FROM `role_permissions` WHERE `role` = 'admin' AND `permission` = 'menu:read';
DROP TABLE IF EXISTS `role_menus`;
DROP TABLE IF EXISTS `menus`;
