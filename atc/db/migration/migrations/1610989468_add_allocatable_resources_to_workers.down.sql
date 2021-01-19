BEGIN;
  ALTER TABLE workers DROP COLUMN allocatable_cpu;
  ALTER TABLE workers DROP COLUMN allocatable_memory;
COMMIT;
