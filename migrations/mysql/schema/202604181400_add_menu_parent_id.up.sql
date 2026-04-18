ALTER TABLE `menus`
  ADD COLUMN `parent_id` BIGINT NULL DEFAULT NULL AFTER `sort`,
  ADD KEY `idx_menus_parent` (`tenant_id`, `parent_id`);
