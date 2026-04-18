ALTER TABLE `system_settings`
ADD COLUMN `tenant_id` VARCHAR(64) NOT NULL DEFAULT 'default',
ADD COLUMN `draft_value` TEXT NOT NULL,
ADD COLUMN `draft_value_type` VARCHAR(16) NOT NULL DEFAULT 'string',
ADD COLUMN `is_published` TINYINT(1) NOT NULL DEFAULT 1,
ADD COLUMN `published_at` DATETIME(3) NULL,
ADD COLUMN `published_by` BIGINT NOT NULL DEFAULT 0,
ADD COLUMN `publish_note` VARCHAR(255) NOT NULL DEFAULT '';

UPDATE `system_settings`
SET `draft_value` = `value`,
    `draft_value_type` = `value_type`
WHERE `draft_value` = '';

ALTER TABLE `system_settings`
-- DROP INDEX `uk_system_settings_key`,
ADD UNIQUE KEY `uk_system_settings_tenant_key` (`tenant_id`, `key`),
ADD KEY `idx_system_settings_tenant` (`tenant_id`);
