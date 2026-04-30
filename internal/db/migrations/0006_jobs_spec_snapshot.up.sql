-- Snapshot the preset spec onto every job row so editing a preset no longer
-- mutates the behavior of queued or in-flight jobs. Backfill from each job's
-- current preset; defensive default for rows whose preset is somehow missing.

ALTER TABLE jobs ADD COLUMN spec_snapshot jsonb;

UPDATE jobs j SET spec_snapshot = p.spec
FROM presets p
WHERE j.preset_id = p.id
  AND j.spec_snapshot IS NULL;

UPDATE jobs SET spec_snapshot = '{}'::jsonb WHERE spec_snapshot IS NULL;

ALTER TABLE jobs ALTER COLUMN spec_snapshot SET NOT NULL;
