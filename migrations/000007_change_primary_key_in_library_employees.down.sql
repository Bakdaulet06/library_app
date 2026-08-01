-- 1. Drop the 3-column Primary Key
ALTER TABLE public.library_employees 
DROP CONSTRAINT library_employees_pkey;

-- 2. Restore the original Primary Key on (library_id, member_id)
ALTER TABLE public.library_employees 
ADD CONSTRAINT library_employees_pkey PRIMARY KEY (library_id, member_id);

-- 3. Re-add the UNIQUE constraint on member_id (restores 1-to-1 member constraint)
ALTER TABLE public.library_employees 
ADD CONSTRAINT library_employees_member_id_key UNIQUE (member_id);

-- 4. Re-add the UNIQUE constraint on library_id (restores 1-to-1 library constraint)
ALTER TABLE public.library_employees 
ADD CONSTRAINT library_employees_library_id_key UNIQUE (library_id);