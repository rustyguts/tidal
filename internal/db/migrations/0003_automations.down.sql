ALTER TABLE jobs DROP CONSTRAINT IF EXISTS jobs_automation_id_fkey;
DROP TABLE IF EXISTS automation_runs;
DROP TABLE IF EXISTS automations;
