-- Recreate the obsolete automations tables for rollback symmetry. Data is not
-- restored — workflows replaced automations in 0005, and any rows lived in
-- the new tables.

CREATE TABLE automations (
	id               uuid PRIMARY KEY,
	name             text NOT NULL UNIQUE,
	enabled          boolean NOT NULL DEFAULT true,
	watch_dir        text NOT NULL,
	glob             text NOT NULL,
	preset_id        uuid NOT NULL REFERENCES presets(id) ON DELETE RESTRICT,
	output_dir       text NOT NULL,
	archive_dir      text NOT NULL,
	poll_interval_ms int NOT NULL DEFAULT 30000,
	debounce_ms      int NOT NULL DEFAULT 5000,
	created_at       timestamptz NOT NULL DEFAULT now(),
	updated_at       timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE automation_runs (
	id            bigserial PRIMARY KEY,
	automation_id uuid NOT NULL REFERENCES automations(id) ON DELETE CASCADE,
	source_path   text NOT NULL,
	job_id        uuid REFERENCES jobs(id) ON DELETE SET NULL,
	outcome       text NOT NULL,
	message       text NOT NULL DEFAULT '',
	occurred_at   timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE jobs ADD COLUMN automation_id uuid REFERENCES automations(id) ON DELETE SET NULL;
ALTER TABLE jobs DROP COLUMN IF EXISTS workflow_id;

DROP INDEX IF EXISTS workflow_runs_dedupe;
DROP INDEX IF EXISTS workflow_runs_wid_time_idx;
DROP TABLE IF EXISTS workflow_runs;
DROP TABLE IF EXISTS workflows;
