INSERT INTO role_menus (tenant_id, role, menu_id, created_at, updated_at) VALUES
  ('default', 'admin', 1, NOW(), NOW()),
  ('default', 'admin', 2, NOW(), NOW()),
  ('default', 'admin', 3, NOW(), NOW()),
  ('default', 'admin', 4, NOW(), NOW()),
  ('default', 'admin', 5, NOW(), NOW()),
  ('default', 'admin', 6, NOW(), NOW())
ON CONFLICT (tenant_id, role, menu_id) DO NOTHING;
