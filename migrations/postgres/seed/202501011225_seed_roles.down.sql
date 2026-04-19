DELETE FROM roles
WHERE tenant_id = 'default' AND code IN ('admin', 'user');
