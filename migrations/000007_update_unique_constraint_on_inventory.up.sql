-- 1. Drop the old unique constraint (no longer needed)
ALTER TABLE book_inventory DROP CONSTRAINT IF EXISTS unique_library_book_shelf;

-- 2. Drop the restrictive primary key
ALTER TABLE book_inventory DROP CONSTRAINT book_inventory_pkey;

-- 3. Create the correct composite primary key
ALTER TABLE book_inventory ADD PRIMARY KEY (library_id, book_id, bookshelf_id);