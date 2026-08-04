-- ============================================
-- Seed data for library DB
-- Run with:
--   docker exec -i library-db-1 psql -U postgres -d library < seed.sql
-- Safe to re-run: clears seeded tables first (TRUNCATE ... CASCADE)
-- Login password for ALL seeded members: Password123!
-- ============================================

BEGIN;

-- Clear existing data (CASCADE handles FK dependencies)
TRUNCATE TABLE
    order_items,
    orders,
    returned_books,
    loans,
    cards,
    library_employees,
    employees,
    book_inventory,
    bookshelves,
    books,
    members,
    genres,
    libraries
    RESTART IDENTITY CASCADE;

-- ============================================
-- Libraries (3)
-- ============================================
INSERT INTO libraries (name, address) VALUES
    ('Central Branch', '123 Main St, Almaty'),
    ('Northside Branch', '45 Abay Ave, Almaty'),
    ('Riverside Branch', '78 Al-Farabi Blvd, Almaty');

-- ============================================
-- Genres
-- ============================================
INSERT INTO genres (name) VALUES
    ('Uncategorized'),
    ('Fiction'),
    ('Non-Fiction'),
    ('Sci-Fi'),
    ('Fantasy'),
    ('Mystery');

-- ============================================
-- Members: 1 admin + 5 clients
-- All passwords = Password123! (bcrypt hash below)
-- ============================================
INSERT INTO members (email, password, role) VALUES
    ('admin@library.com',   '$2b$10$LM9T4aWBKuinwYC.1lW.POPCYbvT2s5f5IvoAh6/e58CKUcFg2Pkq', 'admin'),
    ('alice@test.com',      '$2b$10$LM9T4aWBKuinwYC.1lW.POPCYbvT2s5f5IvoAh6/e58CKUcFg2Pkq', 'client'),
    ('bob@test.com',        '$2b$10$LM9T4aWBKuinwYC.1lW.POPCYbvT2s5f5IvoAh6/e58CKUcFg2Pkq', 'client'),
    ('carol@test.com',      '$2b$10$LM9T4aWBKuinwYC.1lW.POPCYbvT2s5f5IvoAh6/e58CKUcFg2Pkq', 'client'),
    ('dave@test.com',       '$2b$10$LM9T4aWBKuinwYC.1lW.POPCYbvT2s5f5IvoAh6/e58CKUcFg2Pkq', 'client'),
    ('erin@test.com',       '$2b$10$LM9T4aWBKuinwYC.1lW.POPCYbvT2s5f5IvoAh6/e58CKUcFg2Pkq', 'client');

-- ============================================
-- Books (20)
-- ============================================
INSERT INTO books (title, author, isbn, genre_id, price) VALUES
    ('The Hobbit',                    'J.R.R. Tolkien',       '9780547928227', 5, 1000),
    ('The Fellowship of the Ring',    'J.R.R. Tolkien',       '9780547928210', 5, 1000),
    ('Dune',                          'Frank Herbert',        '9780441172719', 4, 1000),
    ('Foundation',                    'Isaac Asimov',         '9780553293357', 4, 1000),
    ('Neuromancer',                   'William Gibson',       '9780441569595', 4, 1000),
    ('1984',                          'George Orwell',        '9780451524935', 2, 1200),
    ('Brave New World',               'Aldous Huxley',        '9780060850524', 2, 1200),
    ('Fahrenheit 451',                'Ray Bradbury',         '9781451673319', 2, 1200),
    ('The Great Gatsby',              'F. Scott Fitzgerald',  '9780743273565', 2, 1200),
    ('To Kill a Mockingbird',         'Harper Lee',            '9780061120084', 2, 1400),
    ('Murder on the Orient Express',  'Agatha Christie',       '9780062693662', 6, 1400),
    ('The Hound of the Baskervilles', 'Arthur Conan Doyle',    '9780451528018', 6, 1400),
    ('Gone Girl',                     'Gillian Flynn',         '9780307588371', 6, 1500),
    ('The Da Vinci Code',             'Dan Brown',             '9780307474278', 6, 900),
    ('Sapiens',                       'Yuval Noah Harari',     '9780062316097', 3, 1800),
    ('Educated',                      'Tara Westover',         '9780399590504', 3, 1800),
    ('Atomic Habits',                 'James Clear',           '9780735211292', 3, 1800),
    ('A Game of Thrones',             'George R.R. Martin',    '9780553103540', 5, 1800),
    ('The Name of the Wind',          'Patrick Rothfuss',      '9780756404741', 5, 1800),
    ('Mistborn',                      'Brandon Sanderson',     '9780765311788', 5, 1800);

-- ============================================
-- Bookshelves (a few per library)
-- ============================================
INSERT INTO bookshelves (library_id, code, capacity, empty_space) VALUES
    (1, 'A-1', 50, 50),
    (1, 'A-2', 50, 50),
    (2, 'A-1', 40, 40),
    (3, 'A-1', 60, 60);

COMMIT;

-- ============================================
-- Quick sanity check
-- ============================================
SELECT 'libraries' AS table_name, COUNT(*) FROM libraries
UNION ALL SELECT 'genres', COUNT(*) FROM genres
UNION ALL SELECT 'members', COUNT(*) FROM members
UNION ALL SELECT 'books', COUNT(*) FROM books
UNION ALL SELECT 'bookshelves', COUNT(*) FROM bookshelves;