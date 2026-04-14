INSERT IGNORE INTO `user_roles` (`user_id`, `role`, `created_at`, `updated_at`)
SELECT `id`, 'admin', NOW(3), NOW(3)
FROM `users`
WHERE `username` = 'admin';
