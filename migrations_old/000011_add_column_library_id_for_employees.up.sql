-- 1. Add library_id column referencing the libraries table
ALTER TABLE employees
ADD COLUMN library_id INT NOT NULL,
ADD CONSTRAINT fk_employees_library 
    FOREIGN KEY (library_id) 
    REFERENCES libraries(id) 
    ON DELETE RESTRICT;

-- 2. Index for faster JOIN queries on library_id
CREATE INDEX idx_employees_library_id ON employees(library_id);