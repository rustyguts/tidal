ALTER TABLE jobs
	DROP COLUMN IF EXISTS cache_path,
	DROP COLUMN IF EXISTS source_move_path;
