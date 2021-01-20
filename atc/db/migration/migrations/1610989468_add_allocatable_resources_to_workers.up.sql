BEGIN;
  ALTER TABLE workers ADD COLUMN allocatable_cpu bigint;
  ALTER TABLE workers ADD COLUMN allocatable_memory bigint;
COMMIT;
