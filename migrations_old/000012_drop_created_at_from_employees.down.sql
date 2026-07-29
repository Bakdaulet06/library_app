-- Restore created_at column if you ever roll back this migration
ALTER TABLE employees
ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();