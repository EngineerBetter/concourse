BEGIN;
  ALTER TABLE containers DROP COLUMN exit_code;
COMMIT;
