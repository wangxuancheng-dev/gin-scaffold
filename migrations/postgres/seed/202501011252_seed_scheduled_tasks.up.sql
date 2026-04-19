-- 示例定时任务：每小时执行一次内置 `artisan ping`（见 internal/console/commands/ping.go）
INSERT INTO scheduled_tasks (
  tenant_id,
  name,
  spec,
  command,
  timeout_sec,
  concurrency_policy,
  enabled,
  created_at,
  updated_at
) VALUES (
  'default',
  'example-hourly-ping',
  '@hourly',
  'artisan ping',
  30,
  'forbid',
  TRUE,
  NOW(),
  NOW()
)
ON CONFLICT (tenant_id, name) DO NOTHING;
