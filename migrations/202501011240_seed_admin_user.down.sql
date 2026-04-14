DELETE ur
FROM `user_roles` ur
JOIN `users` u ON u.id = ur.user_id
WHERE u.username = 'admin' AND ur.role = 'admin';

DELETE FROM `users`
WHERE `username` = 'admin'
  AND `nickname` = 'Administrator';
