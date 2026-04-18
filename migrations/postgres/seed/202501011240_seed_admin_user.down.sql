DELETE FROM user_roles ur
USING users u
WHERE ur.user_id = u.id
  AND u.username = 'admin'
  AND ur.role = 'admin'
  AND u.tenant_id = 'default'
  AND ur.tenant_id = 'default';

DELETE FROM users
WHERE tenant_id = 'default' AND username = 'admin' AND nickname = 'Administrator';
