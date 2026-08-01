-- 1. Drop the UNIQUE constraint on library_id (allows multiple employees per library)
ALTER TABLE public.library_employees 
DROP CONSTRAINT library_employees_library_id_key;

-- 2. Drop the UNIQUE constraint on member_id (allows a member to work multiple roles/libraries)
ALTER TABLE public.library_employees 
DROP CONSTRAINT library_employees_member_id_key;

-- 3. Drop the old Primary Key constraint
ALTER TABLE public.library_employees 
DROP CONSTRAINT library_employees_pkey;

-- 4. Re-create the Primary Key across all 3 columns (library_id, member_id, position)
ALTER TABLE public.library_employees 
ADD CONSTRAINT library_employees_pkey PRIMARY KEY (library_id, member_id, position);