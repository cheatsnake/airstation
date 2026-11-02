package migrations

import (
	"database/sql"
	"fmt"
)

type Migration struct {
	Version int
	Name    string
	Up      func(*sql.Tx) error
	Down    func(*sql.Tx) error
}

var migrations = []Migration{
	{
		Version: 1,
		Name:    "create_main_tables",
		Up: func(tx *sql.Tx) error {
			queries := []string{
				`CREATE TABLE IF NOT EXISTS migrations (
				    version INTEGER PRIMARY KEY,
				    name TEXT NOT NULL,
				    applied_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
				);`,
				`CREATE TABLE IF NOT EXISTS tracks (
                    id TEXT PRIMARY KEY,
                    name TEXT NOT NULL UNIQUE,
                    path TEXT NOT NULL,
                    duration REAL NOT NULL,
                    bitRate INTEGER NOT NULL
                );`,
				`CREATE TABLE IF NOT EXISTS queue (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    track_id TEXT NOT NULL UNIQUE,
                    FOREIGN KEY (track_id) REFERENCES tracks (id)
                );`,
				`CREATE TABLE IF NOT EXISTS playback_history (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    played_at INTEGER NOT NULL,
                    track_name TEXT NOT NULL
                );`,
				`CREATE TABLE IF NOT EXISTS playlist (
                    id TEXT PRIMARY KEY,
                    name TEXT NOT NULL UNIQUE,
                    description TEXT
                );`,
				`CREATE TABLE IF NOT EXISTS playlist_track (
                    playlist_id TEXT NOT NULL,
                    track_id TEXT NOT NULL,
                    position INTEGER NOT NULL,
                    FOREIGN KEY (playlist_id) REFERENCES playlist (id) ON DELETE CASCADE,
                    FOREIGN KEY (track_id) REFERENCES tracks (id),
                    PRIMARY KEY (playlist_id, position),
                    UNIQUE (playlist_id, track_id)
                );`,
				`CREATE TABLE IF NOT EXISTS station_properties (
                    key VARCHAR(100) PRIMARY KEY,
                    value TEXT,
                    created_at INTEGER DEFAULT (strftime('%s', 'now')),
                    updated_at INTEGER DEFAULT (strftime('%s', 'now'))
                );`,
			}

			for _, query := range queries {
				if _, err := tx.Exec(query); err != nil {
					return fmt.Errorf("failed to execute query: %w, query: %s", err, query)
				}
			}
			return nil
		},
	},
	{
		Version: 2,
		Name:    "create_main_indexes",
		Up: func(tx *sql.Tx) error {
			indexes := []string{
				`CREATE INDEX IF NOT EXISTS idx_tracks_name ON tracks (name COLLATE NOCASE);`,
				`CREATE INDEX IF NOT EXISTS idx_playback_history_played_at ON playback_history(played_at);`,
				`CREATE INDEX IF NOT EXISTS idx_playlist_track_ids ON playlist_track (playlist_id, track_id);`,
			}

			for _, query := range indexes {
				if _, err := tx.Exec(query); err != nil {
					return fmt.Errorf("failed to create index: %w, query: %s", err, query)
				}
			}
			return nil
		},
	},
	{
		Version: 3,
		Name:    "add_radio_automation",
		Up: func(tx *sql.Tx) error {
			queries := []string{
				`ALTER TABLE tracks ADD COLUMN kind TEXT NOT NULL DEFAULT 'music';`,
				`ALTER TABLE tracks ADD COLUMN fade_in REAL NOT NULL DEFAULT 0;`,
				`ALTER TABLE tracks ADD COLUMN fade_out REAL NOT NULL DEFAULT 0;`,
				`ALTER TABLE tracks ADD COLUMN duck_db REAL NOT NULL DEFAULT 0;`,
				`ALTER TABLE playback_history ADD COLUMN kind TEXT NOT NULL DEFAULT 'music';`,
				`CREATE TABLE IF NOT EXISTS schedule_rules (
					id TEXT PRIMARY KEY,
					name TEXT NOT NULL,
					kind TEXT NOT NULL,
					every_n_tracks INTEGER NOT NULL DEFAULT 0,
					at_minute INTEGER NOT NULL DEFAULT -1,
					enabled INTEGER NOT NULL DEFAULT 1,
					created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
				);`,
				`CREATE TABLE IF NOT EXISTS scheduler_state (
					key TEXT PRIMARY KEY,
					value TEXT NOT NULL
				);`,
				`CREATE TABLE IF NOT EXISTS voice_overs (
					id TEXT PRIMARY KEY,
					voice_track_id TEXT NOT NULL,
					target TEXT NOT NULL DEFAULT 'next',
					target_track_id TEXT,
					fade_in REAL NOT NULL DEFAULT 1.5,
					fade_out REAL NOT NULL DEFAULT 2.0,
					duck_db REAL NOT NULL DEFAULT -12,
					start_offset REAL NOT NULL DEFAULT 0,
					status TEXT NOT NULL DEFAULT 'pending',
					applied_track_id TEXT,
					mix_path TEXT NOT NULL DEFAULT '',
					mix_id TEXT NOT NULL DEFAULT '',
					created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
					FOREIGN KEY (voice_track_id) REFERENCES tracks (id)
				);`,
				`CREATE INDEX IF NOT EXISTS idx_tracks_kind ON tracks(kind);`,
				`CREATE INDEX IF NOT EXISTS idx_playback_history_kind ON playback_history(kind);`,
				`CREATE INDEX IF NOT EXISTS idx_voice_overs_status ON voice_overs(status, target);`,
			}

			for _, query := range queries {
				if _, err := tx.Exec(query); err != nil {
					return fmt.Errorf("failed to execute query: %w, query: %s", err, query)
				}
			}
			return nil
		},
	},
}
