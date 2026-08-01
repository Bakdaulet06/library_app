-- Down Migration: Restore default 1 for genre_id
ALTER TABLE books ALTER COLUMN genre_id SET DEFAULT 1;