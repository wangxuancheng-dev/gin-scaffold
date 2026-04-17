INSERT INTO `users` (`tenant_id`, `username`, `password`, `nickname`, `created_at`, `updated_at`)
SELECT 'default', 'admin', '$2a$10$rLrUUz3msxEX4F0khc8Ane/.UhBlU3Jib02NKP09U3S8sAvhuODnG', 'Administrator', NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM `users` WHERE `tenant_id` = 'default' AND `username` = 'admin'
);

INSERT IGNORE INTO `user_roles` (`tenant_id`, `user_id`, `role`, `created_at`, `updated_at`)
SELECT 'default', `id`, 'admin', NOW(), NOW()
FROM `users`
WHERE `tenant_id` = 'default' AND `username` = 'admin';

-- Default admin password (for first login only): Admin@123456
-- IMPORTANT: change password immediately after first login.
