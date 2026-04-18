DELETE FROM role_permissions WHERE tenant_id = 'default' AND role = 'admin' AND permission IN ('announcement:read', 'announcement:write');
