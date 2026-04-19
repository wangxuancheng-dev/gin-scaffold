DELETE ur
FROM `user_roles` ur
JOIN `users` u ON u.id = ur.user_id
WHERE u.tenant_id = 'default' AND ur.tenant_id = 'default'
  AND u.username = 'admin' AND ur.role = 'admin';
