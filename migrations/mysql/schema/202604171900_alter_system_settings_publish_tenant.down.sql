ALTER TABLE `system_settings`
DROP INDEX `uk_system_settings_tenant_key`,
DROP INDEX `idx_system_settings_tenant`,
ADD UNIQUE KEY `uk_system_settings_key` (`key`);

ALTER TABLE `system_settings`
DROP COLUMN `publish_note`,
DROP COLUMN `published_by`,
DROP COLUMN `published_at`,
DROP COLUMN `is_published`,
DROP COLUMN `draft_value_type`,
DROP COLUMN `draft_value`,
DROP COLUMN `tenant_id`;
