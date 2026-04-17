DELETE FROM role_permissions WHERE role = 'admin' AND permission IN ('announcement:read', 'announcement:write');
