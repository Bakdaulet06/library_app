-- Undo loan changes
ALTER TABLE loans 
DROP COLUMN IF EXISTS returned_library_id,
DROP COLUMN IF EXISTS borrowed_library_id;

-- Undo book changes
ALTER TABLE books 
DROP COLUMN IF EXISTS genre_id,
DROP COLUMN IF EXISTS library_id;

-- Drop lookup tables
DROP TABLE IF EXISTS genres;
DROP TABLE IF EXISTS libraries;