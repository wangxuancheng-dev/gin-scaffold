DELETE FROM `role_menus`
WHERE `role` = 'admin' AND `menu_id` IN (1, 2, 3);

DELETE FROM `menus`
WHERE `id` IN (1, 2, 3);
