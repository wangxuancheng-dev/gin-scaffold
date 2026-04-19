INSERT IGNORE INTO `user_roles` (`tenant_id`, `user_id`, `role`, `created_at`, `updated_at`)
SELECT 'default', `id`, 'admin', NOW(), NOW()
FROM `users`
WHERE `tenant_id` = 'default' AND `username` = 'admin';
