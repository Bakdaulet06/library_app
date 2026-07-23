ALTER TABLE loans
RENAME COLUMN loan_date TO borrowed_at;

ALTER TABLE loans
RENAME COLUMN return_date TO returned_at;