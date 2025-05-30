package sqlite

import (
	"database/sql"
	"fmt"
	"log/slog"
	"sync"

	_ "modernc.org/sqlite"
)

type Instance struct {
	TrackStore
	QueueStore
	PlaybackStore
	PlaylistStore

	db    *sql.DB
	log   *slog.Logger
	mutex sync.Mutex
}

func New(dbPath string, log *slog.Logger) (*Instance, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(1)
	log.Info("Sqlite database connected.")

	err = createTables(db)
	if err != nil {
		return nil, err
	}

	err = createIndexes(db)
	if err != nil {
		return nil, err
	}

	instance := &Instance{
		db:  db,
		log: log,
	}

	instance.TrackStore = NewTrackStore(db, &instance.mutex)
	instance.QueueStore = NewQueueStore(db, &instance.mutex)
	instance.PlaybackStore = NewPlaybackStore(db, &instance.mutex)
	instance.PlaylistStore = NewPlaylistStore(db, &instance.mutex)

	return instance, nil
}

func (ins *Instance) Close() error {
	ins.mutex.Lock()
	defer ins.mutex.Unlock()
	return ins.db.Close()
}

func createTables(db *sql.DB) error {
	tracksTable := `
		CREATE TABLE IF NOT EXISTS tracks (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			path TEXT NOT NULL,
			duration REAL NOT NULL,
			bitRate INTEGER NOT NULL
		);`

	queueTable := `
		CREATE TABLE IF NOT EXISTS queue (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			track_id TEXT NOT NULL UNIQUE,
			FOREIGN KEY (track_id) REFERENCES tracks (id)
		);`

	playbackHistoryTable := `
		CREATE TABLE IF NOT EXISTS playback_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			played_at INTEGER NOT NULL,
			track_name TEXT NOT NULL
		);`

	playlistTable := `
		CREATE TABLE IF NOT EXISTS playlist (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT
		);`

	playlistTrackTable := `
		CREATE TABLE IF NOT EXISTS playlist_track (
			playlist_id TEXT NOT NULL,
			track_id TEXT NOT NULL,
			FOREIGN KEY (playlist_id) REFERENCES playlist (id),
			FOREIGN KEY (track_id) REFERENCES tracks (id),
			UNIQUE (playlist_id, track_id)
		);`

	_, err := db.Exec(tracksTable)
	if err != nil {
		return fmt.Errorf("failed to create tracks table: %w", err)
	}

	_, err = db.Exec(queueTable)
	if err != nil {
		return fmt.Errorf("failed to create queue table: %w", err)
	}

	_, err = db.Exec(playbackHistoryTable)
	if err != nil {
		return fmt.Errorf("failed to create table for playback history: %w", err)
	}

	_, err = db.Exec(playlistTable)
	if err != nil {
		return fmt.Errorf("failed to create playlist table: %w", err)
	}

	_, err = db.Exec(playlistTrackTable)
	if err != nil {
		return fmt.Errorf("failed to create playlist track table: %w", err)
	}

	return nil
}

func createIndexes(db *sql.DB) error {
	indexQuery := `CREATE INDEX IF NOT EXISTS idx_tracks_name ON tracks (name COLLATE NOCASE);`
	playedAtIndexQuery := `CREATE INDEX IF NOT EXISTS idx_playback_history_played_at ON playback_history(played_at);`
	playlistTrackIndexQuery := `CREATE INDEX IF NOT EXISTS idx_playlist_track_ids ON playlist_track (playlist_id, track_id);`

	_, err := db.Exec(indexQuery)
	if err != nil {
		return fmt.Errorf("failed to create index on tracks.name: %w", err)
	}

	_, err = db.Exec(playedAtIndexQuery)
	if err != nil {
		return fmt.Errorf("failed to create index on playback_history.played_at: %w", err)
	}

	_, err = db.Exec(playlistTrackIndexQuery)
	if err != nil {
		return fmt.Errorf("failed to create index for playlist_track: %w", err)
	}

	return nil
}
