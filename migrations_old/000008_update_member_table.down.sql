-- Drop dependent tables first (reverse order)
DROP TABLE IF EXISTS library_employees;

-- Remove added column from members
ALTER TABLE members 
DROP COLUMN IF EXISTS password,
DROP COLUMN IF EXISTS role;