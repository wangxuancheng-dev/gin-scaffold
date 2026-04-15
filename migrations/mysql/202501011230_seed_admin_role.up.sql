INSERT IGNORE INTO `user_roles` (`user_id`, `role`, `created_at`, `updated_at`)
SELECT `id`, 'admin', NOW(), NOW()
FROM `users`
WHERE `username` = 'admin';
