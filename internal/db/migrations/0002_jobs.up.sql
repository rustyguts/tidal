CREATE TABLE jobs (
	id            uuid PRIMARY KEY,
	asynq_id      text,
	preset_id     uuid NOT NULL REFERENCES presets(id) ON DELETE RESTRICT,
	source_path   text NOT NULL,
	output_path   text NOT NULL,
	status        text NOT NULL,
	k8s_job_name  text,
	automation_id uuid,
	progress      jsonb NOT NULL DEFAULT '{}'::jsonb,
	error         text NOT NULL DEFAULT '',
	created_at    timestamptz NOT NULL DEFAULT now(),
	started_at    timestamptz,
	finished_at   timestamptz
);

CREATE INDEX jobs_status_created_idx ON jobs (status, created_at DESC);
CREATE INDEX jobs_preset_idx         ON jobs (preset_id);
CREATE INDEX jobs_automation_idx     ON jobs (automation_id) WHERE automation_id IS NOT NULL;

CREATE TABLE job_logs (
	job_id     uuid NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
	seq        bigint NOT NULL,
	stream     text NOT NULL,
	line       text NOT NULL,
	emitted_at timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (job_id, seq)
);

CREATE INDEX job_logs_emitted_idx ON job_logs (job_id, emitted_at);
