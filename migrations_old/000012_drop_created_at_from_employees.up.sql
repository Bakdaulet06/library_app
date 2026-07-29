-- Remove created_at column from employees table
ALTER TABLE employees
DROP COLUMN IF EXISTS created_at;