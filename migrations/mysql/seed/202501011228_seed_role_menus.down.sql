DELETE FROM `role_menus`
WHERE `tenant_id` = 'default' AND `role` = 'admin' AND `menu_id` IN (1, 2, 3, 4, 5, 6);
