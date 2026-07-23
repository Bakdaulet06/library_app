-- 1. Create the libraries table
CREATE TABLE IF NOT EXISTS libraries (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seed default central library (ID: 1)
INSERT INTO libraries (id, name, address) 
VALUES (1, 'Central Branch', '123 Main St') 
ON CONFLICT (id) DO NOTHING;

-- 2. Create the genres table
CREATE TABLE IF NOT EXISTS genres (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);

-- Seed default allowed genres
INSERT INTO genres (id, name) VALUES 
    (1, 'Uncategorized'),
    (2, 'Fiction'),
    (3, 'Non-Fiction'),
    (4, 'Sci-Fi'),
    (5, 'Fantasy'),
    (6, 'Mystery')
ON CONFLICT (id) DO NOTHING;

-- 3. Update books table with Foreign Keys
-- Standardizing existing books to default library (1) and default genre (1)
ALTER TABLE books 
ADD COLUMN IF NOT EXISTS library_id INT REFERENCES libraries(id) DEFAULT 1,
ADD COLUMN IF NOT EXISTS genre_id INT REFERENCES genres(id) DEFAULT 1;

-- 4. Update loans table for multi-branch tracking
ALTER TABLE loans 
ADD COLUMN IF NOT EXISTS borrowed_library_id INT REFERENCES libraries(id),
ADD COLUMN IF NOT EXISTS returned_library_id INT REFERENCES libraries(id);