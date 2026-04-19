DELETE FROM scheduled_tasks
WHERE tenant_id = 'default'
  AND name = 'example-hourly-ping';
