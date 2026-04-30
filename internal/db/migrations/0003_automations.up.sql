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
	outcome       text NOT NULL,            -- matched|enqueued|skipped_dupe|error|archived
	message       text NOT NULL DEFAULT '',
	occurred_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX automation_runs_aid_time_idx ON automation_runs (automation_id, occurred_at DESC);
CREATE UNIQUE INDEX automation_runs_dedupe ON automation_runs (automation_id, source_path)
	WHERE outcome = 'enqueued';

-- Wire jobs.automation_id to the new table.
ALTER TABLE jobs
	ADD CONSTRAINT jobs_automation_id_fkey
	FOREIGN KEY (automation_id) REFERENCES automations(id) ON DELETE SET NULL;
