-- 000003_create_book_inventory.up.sql

-- 1. Create junction table for Many-to-Many relationship
CREATE TABLE IF NOT EXISTS book_inventory (
    library_id INT NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
    book_id INT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    available_copies INT NOT NULL DEFAULT 1 CHECK (available_copies >= 0),
    PRIMARY KEY (library_id, book_id)
);

-- 2. Remove obsolete columns from books table
ALTER TABLE books DROP COLUMN IF EXISTS library_id;
ALTER TABLE books DROP COLUMN IF EXISTS available_copies;