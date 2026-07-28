-- 1. Drop the index
DROP INDEX IF EXISTS idx_employees_library_id;

-- 2. Drop the foreign key constraint and column
ALTER TABLE employees
DROP CONSTRAINT IF EXISTS fk_employees_library,
DROP COLUMN IF EXISTS library_id;