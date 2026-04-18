DELETE FROM role_menus WHERE tenant_id = 'default' AND role = 'admin' AND menu_id = 6;
DELETE FROM menus WHERE tenant_id = 'default' AND id = 6 AND perm_code = 'sys:config:read';
DELETE FROM role_permissions WHERE tenant_id = 'default' AND role = 'admin' AND permission IN ('sys:config:read', 'sys:config:write');
