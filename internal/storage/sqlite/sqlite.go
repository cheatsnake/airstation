package sqlite

import (
	"database/sql"
	"fmt"
	"log/slog"
	"sync"

	"github.com/cheatsnake/airstation/internal/storage/sqlite/migrations"
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

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	db.SetMaxOpenConns(1)
	_, _ = db.Exec("PRAGMA journal_mode = WAL")
	_, _ = db.Exec("PRAGMA synchronous = NORMAL")

	log.Info("Sqlite database connected")

	err = migrations.RunMigrations(db, log)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
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

	if _, err := ins.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		ins.mutex.Unlock()
		return fmt.Errorf("failed to run wal checkpoint: %w", err)
	}

	err := ins.db.Close()
	ins.mutex.Unlock()

	return err
}
