-- Migration: split accounts_payable.status into two columns
--
-- Before: accounts_payable.status held both lifecycle values (PENDING, APPROVED, …)
--         and payment-state values (PAID, PARTIAL).
--
-- After:
--   status      TEXT  — PayableStatus lifecycle : draft | pending | approved | rejected | void
--   paid_status TEXT  — PaidStatus payment state: unpaid | partial | paid | refunded

-- 1. Add the new paid_status column (defaults to 'unpaid' for existing rows)
ALTER TABLE accounts_payable
    ADD COLUMN IF NOT EXISTS paid_status TEXT NOT NULL DEFAULT 'unpaid';

-- 2. Migrate any rows whose current status is a payment-state value.
--    These were written by updateAPBalance before the refactor.
UPDATE accounts_payable
   SET paid_status = LOWER(status),
       status      = 'pending'
 WHERE LOWER(status) IN ('paid', 'partial', 'unpaid');

-- 3. Lower-case remaining lifecycle status values that are still uppercase
--    (written before the PayableStatus → lowercase refactor).
UPDATE accounts_payable
   SET status = LOWER(status)
 WHERE status != LOWER(status);

-- 4. If the payable_status enum type exists, drop it now that the column is TEXT.
--    Run only if your DB has this enum; comment out if it does not exist.
-- ALTER TABLE accounts_payable ALTER COLUMN status TYPE TEXT;
-- DROP TYPE IF EXISTS payable_status;
