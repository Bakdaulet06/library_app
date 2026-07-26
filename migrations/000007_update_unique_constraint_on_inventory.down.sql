-- 1. Drop the composite primary key (library_id, book_id, bookshelf_id)
ALTER TABLE book_inventory DROP CONSTRAINT book_inventory_pkey;

-- 2. Restore the original restrictive primary key (library_id, book_id)
ALTER TABLE book_inventory ADD PRIMARY KEY (library_id, book_id);

-- 3. Restore the separate unique constraint
ALTER TABLE book_inventory 
ADD CONSTRAINT unique_library_book_shelf UNIQUE (library_id, book_id, bookshelf_id);