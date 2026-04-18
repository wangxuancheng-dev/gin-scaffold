DELETE FROM role_menus WHERE tenant_id = 'default' AND role = 'admin' AND menu_id = 5;
DELETE FROM menus WHERE tenant_id = 'default' AND id = 5 AND perm_code = 'audit:read';
DELETE FROM role_permissions WHERE tenant_id = 'default' AND role = 'admin' AND permission = 'audit:read';
