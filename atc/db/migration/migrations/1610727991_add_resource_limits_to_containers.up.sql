BEGIN;
  ALTER TABLE containers ADD COLUMN meta_cpu_limit bigint;
  ALTER TABLE containers ADD COLUMN meta_memory_limit bigint;
COMMIT;
