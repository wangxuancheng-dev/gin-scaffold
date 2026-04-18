DROP INDEX IF EXISTS idx_menus_parent;

ALTER TABLE menus
  DROP COLUMN IF EXISTS parent_id;
