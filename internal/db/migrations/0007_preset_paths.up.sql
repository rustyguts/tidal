ALTER TABLE presets
    ADD COLUMN output_path       text NOT NULL DEFAULT '',
    ADD COLUMN cache_path        text NOT NULL DEFAULT '',
    ADD COLUMN source_move_path  text NOT NULL DEFAULT '';
