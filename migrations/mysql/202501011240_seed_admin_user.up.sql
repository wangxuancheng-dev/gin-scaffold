INSERT INTO `users` (`username`, `password`, `nickname`, `created_at`, `updated_at`)
SELECT 'admin', '$2a$10$rLrUUz3msxEX4F0khc8Ane/.UhBlU3Jib02NKP09U3S8sAvhuODnG', 'Administrator', NOW(3), NOW(3)
WHERE NOT EXISTS (
  SELECT 1 FROM `users` WHERE `username` = 'admin'
);

INSERT IGNORE INTO `user_roles` (`user_id`, `role`, `created_at`, `updated_at`)
SELECT `id`, 'admin', NOW(3), NOW(3)
FROM `users`
WHERE `username` = 'admin';

-- Default admin password (for first login only): Admin@123456
-- IMPORTANT: change password immediately after first login.
