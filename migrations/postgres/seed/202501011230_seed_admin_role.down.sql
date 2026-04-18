DELETE FROM user_roles ur
USING users u
WHERE ur.user_id = u.id
  AND u.tenant_id = 'default'
  AND u.username = 'admin'
  AND ur.role = 'admin'
  AND ur.tenant_id = 'default';
