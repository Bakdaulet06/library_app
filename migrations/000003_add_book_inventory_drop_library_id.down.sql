-- 000003_create_book_inventory.down.sql

-- 1. Restore the dropped columns to the books table
ALTER TABLE books ADD COLUMN IF NOT EXISTS library_id INT REFERENCES libraries(id) ON DELETE SET NULL;
ALTER TABLE books ADD COLUMN IF NOT EXISTS available_copies INT NOT NULL DEFAULT 1 CHECK (available_copies >= 0);

-- 2. Drop the junction table
DROP TABLE IF EXISTS book_inventory;