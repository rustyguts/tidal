ALTER TABLE jobs
	ADD COLUMN cache_path       text NOT NULL DEFAULT '',
	ADD COLUMN source_move_path text NOT NULL DEFAULT '';
