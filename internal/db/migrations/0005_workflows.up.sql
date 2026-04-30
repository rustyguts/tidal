CREATE TABLE workflows (
	id                  uuid PRIMARY KEY,
	name                text NOT NULL UNIQUE,
	enabled             boolean NOT NULL DEFAULT true,
	trigger             jsonb NOT NULL,
	actions             jsonb NOT NULL,
	poll_interval_ms    int NOT NULL DEFAULT 30000,
	stable_threshold_ms int NOT NULL DEFAULT 60000,
	runs_count          bigint NOT NULL DEFAULT 0,
	success_count       bigint NOT NULL DEFAULT 0,
	last_run_at         timestamptz,
	created_at          timestamptz NOT NULL DEFAULT now(),
	updated_at          timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE workflow_runs (
	id          bigserial PRIMARY KEY,
	workflow_id uuid NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
	source_path text NOT NULL,
	job_id      uuid REFERENCES jobs(id) ON DELETE SET NULL,
	outcome     text NOT NULL,
	message     text NOT NULL DEFAULT '',
	occurred_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX workflow_runs_wid_time_idx ON workflow_runs (workflow_id, occurred_at DESC);
CREATE UNIQUE INDEX workflow_runs_dedupe ON workflow_runs (workflow_id, source_path)
	WHERE outcome IN ('enqueued', 'triggered');

ALTER TABLE jobs ADD COLUMN workflow_id uuid REFERENCES workflows(id) ON DELETE SET NULL;
ALTER TABLE jobs DROP CONSTRAINT IF EXISTS jobs_automation_id_fkey;
ALTER TABLE jobs DROP COLUMN IF EXISTS automation_id;

DROP TABLE IF EXISTS automation_runs;
DROP TABLE IF EXISTS automations;
