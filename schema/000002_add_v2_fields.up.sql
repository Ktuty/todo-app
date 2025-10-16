-- Добавляем новые поля для v2 API
ALTER TABLE todo_lists
    ADD COLUMN IF NOT EXISTS archived boolean not null default false,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone not null default current_timestamp,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone not null default current_timestamp,
    ADD COLUMN IF NOT EXISTS color varchar(50) default '#000000',
    ADD COLUMN IF NOT EXISTS priority integer not null default 0;

ALTER TABLE todo_items
    ADD COLUMN IF NOT EXISTS archived boolean not null default false,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone not null default current_timestamp,
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone not null default current_timestamp;

-- Создаем индексы для улучшения производительности
CREATE INDEX IF NOT EXISTS idx_todo_lists_archived ON todo_lists(archived);
CREATE INDEX IF NOT EXISTS idx_todo_lists_created_at ON todo_lists(created_at);
CREATE INDEX IF NOT EXISTS idx_todo_items_archived ON todo_items(archived);
CREATE INDEX IF NOT EXISTS idx_todo_items_done ON todo_items(done);
CREATE INDEX IF NOT EXISTS idx_todo_items_created_at ON todo_items(created_at);