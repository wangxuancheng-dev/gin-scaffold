-- 优化按 enabled 拉取并按 id 排序的调度查询（如 ListEnabled）。
ALTER TABLE `scheduled_tasks` ADD INDEX `idx_scheduled_tasks_enabled_id` (`enabled`, `id`);
