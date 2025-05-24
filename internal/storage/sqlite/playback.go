package sqlite

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/cheatsnake/airstation/internal/playback"
)

type PlaybackStore struct {
	db    *sql.DB
	mutex *sync.Mutex
}

func NewPlaybackStore(db *sql.DB, mutex *sync.Mutex) PlaybackStore {
	return PlaybackStore{
		db:    db,
		mutex: mutex,
	}
}

func (ps *PlaybackStore) AddPlaybackHistory(playedAt int64, trackName string) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	query := `INSERT INTO playback_history (played_at, track_name) VALUES (?, ?)`

	_, err := ps.db.Exec(query, playedAt, trackName)
	if err != nil {
		return fmt.Errorf("failed to insert playback entry: %v", err)
	}

	return nil
}

func (ps *PlaybackStore) RecentPlaybackHistory(limit int) ([]*playback.History, error) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	query := `
		SELECT id, played_at, track_name 
		FROM playback_history 
		ORDER BY played_at DESC`

	query += fmt.Sprintf(" LIMIT %d", limit)

	rows, err := ps.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*playback.History
	for rows.Next() {
		var item playback.History
		if err := rows.Scan(&item.ID, &item.PlayedAt, &item.TrackName); err != nil {
			return nil, err
		}
		history = append(history, &item)
	}
	return history, nil
}

func (ps *PlaybackStore) DeleteOldPlaybackHistory() (int64, error) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	query := `
		DELETE FROM playback_history 
		WHERE played_at < (strftime('%s', 'now') - 30 * 24 * 60 * 60)`

	result, err := ps.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old entries: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %v", err)
	}

	return rowsAffected, nil
}
