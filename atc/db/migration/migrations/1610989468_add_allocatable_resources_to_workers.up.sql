BEGIN;
  ALTER TABLE workers ADD COLUMN allocatable_cpu integer;
  ALTER TABLE workers ADD COLUMN allocatable_memory integer;
COMMIT;
