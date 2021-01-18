BEGIN;
  ALTER TABLE containers ADD COLUMN meta_memory_limit bigint;
COMMIT;
