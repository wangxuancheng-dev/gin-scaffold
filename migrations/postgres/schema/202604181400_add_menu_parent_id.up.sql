ALTER TABLE menus
  ADD COLUMN IF NOT EXISTS parent_id BIGINT NULL;

CREATE INDEX IF NOT EXISTS idx_menus_parent ON menus (tenant_id, parent_id);
