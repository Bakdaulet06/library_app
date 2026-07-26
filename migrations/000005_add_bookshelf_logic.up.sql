-- 1. Create Bookshelves table without the section column
CREATE TABLE bookshelves (
    id SERIAL PRIMARY KEY,
    library_id INT NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,            -- e.g., "A-101", "SHELF-01"
    capacity INT NOT NULL DEFAULT 50,      -- Total max capacity
    empty_space INT NOT NULL DEFAULT 50,   -- Starts equal to capacity!
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT chk_empty_space_non_negative CHECK (empty_space >= 0),
    CONSTRAINT chk_empty_space_not_exceed_capacity CHECK (empty_space <= capacity),
    UNIQUE(library_id, code)               -- Shelf codes are unique per library branch
);

-- 2. Link book_inventory to bookshelves
ALTER TABLE book_inventory 
ADD COLUMN IF NOT EXISTS bookshelf_id INT REFERENCES bookshelves(id) ON DELETE SET NULL;