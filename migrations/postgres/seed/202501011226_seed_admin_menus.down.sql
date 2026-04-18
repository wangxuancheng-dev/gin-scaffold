DELETE FROM role_menus
WHERE tenant_id = 'default' AND role = 'admin' AND menu_id IN (1, 2, 3, 4);

DELETE FROM menus
WHERE tenant_id = 'default' AND id IN (1, 2, 3, 4);
