ALTER TABLE `role_menus`
DROP INDEX `uk_role_tenant_menu`,
DROP INDEX `idx_role_menus_tenant`,
ADD UNIQUE KEY `uk_role_menu` (`role`, `menu_id`),
DROP COLUMN `tenant_id`;

ALTER TABLE `menus`
DROP INDEX `uk_menus_tenant_path`,
DROP INDEX `idx_menus_tenant`,
ADD UNIQUE KEY `uk_menus_path` (`path`),
DROP COLUMN `tenant_id`;

ALTER TABLE `role_permissions`
DROP INDEX `uk_role_tenant_permission`,
DROP INDEX `idx_role_permissions_tenant`,
ADD UNIQUE KEY `uk_role_permission` (`role`, `permission`),
DROP COLUMN `tenant_id`;

ALTER TABLE `user_roles`
DROP INDEX `uk_user_tenant_role`,
DROP INDEX `idx_user_roles_tenant`,
ADD UNIQUE KEY `uk_user_role` (`user_id`, `role`),
DROP COLUMN `tenant_id`;

ALTER TABLE `roles`
DROP INDEX `uk_roles_tenant_code`,
DROP INDEX `idx_roles_tenant`,
ADD UNIQUE KEY `uk_roles_code` (`code`),
DROP COLUMN `tenant_id`;
