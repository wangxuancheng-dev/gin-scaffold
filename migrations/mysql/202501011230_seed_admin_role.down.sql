DELETE ur
FROM `user_roles` ur
JOIN `users` u ON u.id = ur.user_id
WHERE u.username = 'admin' AND ur.role = 'admin';
