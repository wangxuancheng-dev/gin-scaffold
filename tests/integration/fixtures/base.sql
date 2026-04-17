-- Integration fixture users for export-related tests.
INSERT INTO `users` (`username`, `password`, `nickname`, `created_at`, `updated_at`)
VALUES
  ('it_user_1', '$2a$10$rLrUUz3msxEX4F0khc8Ane/.UhBlU3Jib02NKP09U3S8sAvhuODnG', 'Integration User 1', NOW(), NOW()),
  ('it_user_2', '$2a$10$rLrUUz3msxEX4F0khc8Ane/.UhBlU3Jib02NKP09U3S8sAvhuODnG', 'Integration User 2', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  `nickname` = VALUES(`nickname`),
  `updated_at` = NOW();

INSERT IGNORE INTO `user_roles` (`tenant_id`, `user_id`, `role`, `created_at`, `updated_at`)
SELECT 'default', `id`, 'user', NOW(), NOW() FROM `users` WHERE `username` IN ('it_user_1', 'it_user_2');
