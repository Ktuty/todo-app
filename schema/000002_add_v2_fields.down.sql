-- Откат изменений
ALTER TABLE todo_lists
DROP COLUMN IF EXISTS archived,
DROP COLUMN IF EXISTS created_at,
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS color,
DROP COLUMN IF EXISTS priority;

ALTER TABLE todo_items
DROP COLUMN IF EXISTS archived,
DROP COLUMN IF EXISTS created_at,
DROP COLUMN IF EXISTS updated_at;

DROP INDEX IF EXISTS idx_todo_lists_archived;
DROP INDEX IF EXISTS idx_todo_lists_created_at;
DROP INDEX IF EXISTS idx_todo_items_archived;
DROP INDEX IF EXISTS idx_todo_items_done;
DROP INDEX IF EXISTS idx_todo_items_created_at;