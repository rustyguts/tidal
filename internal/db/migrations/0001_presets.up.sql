CREATE TABLE presets (
	id          uuid PRIMARY KEY,
	name        text NOT NULL UNIQUE,
	description text NOT NULL DEFAULT '',
	builtin     boolean NOT NULL DEFAULT false,
	spec        jsonb NOT NULL,
	created_at  timestamptz NOT NULL DEFAULT now(),
	updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX presets_builtin_idx ON presets (builtin);

-- Marker table tracking which builtin preset *names* have been planted.
-- A user may delete a builtin preset; we never resurrect it on restart.
-- POST /api/presets/restore-defaults clears marker rows for missing names
-- to allow re-seeding.
CREATE TABLE seeded_presets (
	name      text PRIMARY KEY,
	seeded_at timestamptz NOT NULL DEFAULT now()
);
