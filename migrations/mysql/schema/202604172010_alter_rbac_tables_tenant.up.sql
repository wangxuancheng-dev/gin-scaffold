ALTER TABLE `roles`
-- ADD COLUMN `tenant_id` VARCHAR(64) NOT NULL DEFAULT 'default',
-- DROP INDEX `uk_roles_code`,
-- ADD UNIQUE KEY `uk_roles_tenant_code` (`tenant_id`, `code`),
-- ADD KEY `idx_roles_tenant` (`tenant_id`);

ALTER TABLE `user_roles`
-- ADD COLUMN `tenant_id` VARCHAR(64) NOT NULL DEFAULT 'default',
-- DROP INDEX `uk_user_role`,
-- ADD UNIQUE KEY `uk_user_tenant_role` (`tenant_id`, `user_id`, `role`),
-- ADD KEY `idx_user_roles_tenant` (`tenant_id`);

ALTER TABLE `role_permissions`
-- ADD COLUMN `tenant_id` VARCHAR(64) NOT NULL DEFAULT 'default',
-- DROP INDEX `uk_role_permission`,
-- ADD UNIQUE KEY `uk_role_tenant_permission` (`tenant_id`, `role`, `permission`),
-- ADD KEY `idx_role_permissions_tenant` (`tenant_id`);

ALTER TABLE `menus`
-- ADD COLUMN `tenant_id` VARCHAR(64) NOT NULL DEFAULT 'default',
-- DROP INDEX `uk_menus_path`,
-- ADD UNIQUE KEY `uk_menus_tenant_path` (`tenant_id`, `path`),
-- ADD KEY `idx_menus_tenant` (`tenant_id`);

ALTER TABLE `role_menus`
-- ADD COLUMN `tenant_id` VARCHAR(64) NOT NULL DEFAULT 'default',
-- DROP INDEX `uk_role_menu`,
-- ADD UNIQUE KEY `uk_role_tenant_menu` (`tenant_id`, `role`, `menu_id`),
-- ADD KEY `idx_role_menus_tenant` (`tenant_id`);
