ALTER TABLE loans
RENAME COLUMN borrowed_at TO loan_date;

ALTER TABLE loans
RENAME COLUMN returned_at TO return_date;