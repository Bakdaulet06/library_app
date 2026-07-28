-- 3. Update / Recreate Members Table
ALTER TABLE members 
ADD COLUMN IF NOT EXISTS password VARCHAR(255) NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS role VARCHAR(255) NOT NULL DEFAULT 'client';
-- 5. Separate Library-Employee Assignment (1-to-1)
-- Links a member with 'employee' role to a specific library
CREATE TABLE IF NOT EXISTS library_employees (
    library_id INT NOT NULL UNIQUE REFERENCES libraries(id) ON DELETE CASCADE,
    member_id INT NOT NULL UNIQUE REFERENCES members(id) ON DELETE CASCADE,
    PRIMARY KEY (library_id, member_id)
);