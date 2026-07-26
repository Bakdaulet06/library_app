-- 1. Remove the foreign key reference from book_inventory table
ALTER TABLE book_inventory DROP COLUMN IF EXISTS bookshelf_id;

-- 2. Drop the bookshelves table
DROP TABLE IF EXISTS bookshelves;