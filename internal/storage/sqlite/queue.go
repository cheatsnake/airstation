package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/cheatsnake/airstation/internal/track"
)

type QueueStore struct {
	db    *sql.DB
	mutex *sync.Mutex
}

func NewQueueStore(db *sql.DB, mutex *sync.Mutex) QueueStore {
	return QueueStore{
		db:    db,
		mutex: mutex,
	}
}

func (qs *QueueStore) Queue() ([]*track.Track, error) {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	tracks := make([]*track.Track, 0, 10)

	query := `
		SELECT ` + trackColumnsAliased + `
		FROM tracks t
		JOIN queue q ON t.id = q.track_id
		ORDER BY q.id ASC`
	rows, err := qs.db.Query(query)
	if err != nil {
		return tracks, fmt.Errorf("failed to query tracks in queue: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		t, err := scanTrack(rows)
		if err != nil {
			return tracks, fmt.Errorf("failed to scan track: %w", err)
		}
		tracks = append(tracks, t)
	}

	if err = rows.Err(); err != nil {
		return tracks, fmt.Errorf("error iterating over rows: %w", err)
	}

	return tracks, nil
}

func (qs *QueueStore) AddToQueue(tracks []*track.Track) error {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	query := `
			INSERT INTO queue (track_id)
			VALUES (?)
			ON CONFLICT (track_id) DO NOTHING
		`

	for _, track := range tracks {
		_, err := qs.db.Exec(query, track.ID)
		if err != nil {
			return fmt.Errorf("failed to add track to queue: %w", err)
		}
	}

	return nil
}

func (qs *QueueStore) RemoveFromQueue(trackIDs []string) error {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	query := `DELETE FROM queue WHERE track_id = ?`
	for _, id := range trackIDs {
		_, err := qs.db.Exec(query, id)
		if err != nil {
			return fmt.Errorf("failed to remove track from queue: %w", err)
		}
	}

	return nil
}

func (qs *QueueStore) ReorderQueue(trackIDs []string) error {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	_, err := qs.db.Exec(`DELETE FROM queue`)
	if err != nil {
		return fmt.Errorf("failed to clear queue: %w", err)
	}

	query := `INSERT INTO queue (track_id) VALUES (?)`
	for _, id := range trackIDs {
		_, err := qs.db.Exec(query, id)
		if err != nil {
			return fmt.Errorf("failed to reorder queue: %w", err)
		}
	}

	return nil
}

func (qs *QueueStore) CurrentAndNextTrack() (*track.Track, *track.Track, error) {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	query := `
	SELECT ` + trackColumnsAliased + `
	FROM tracks t
	JOIN queue q ON t.id = q.track_id
	ORDER BY q.id ASC
	LIMIT 2`
	rows, err := qs.db.Query(query)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query first and second tracks: %w", err)
	}
	defer rows.Close()

	firstTrack := (*track.Track)(nil)
	secondTrack := (*track.Track)(nil)

	for rows.Next() {
		t, err := scanTrack(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan track: %w", err)
		}
		if firstTrack == nil {
			firstTrack = t
		} else if secondTrack == nil {
			secondTrack = t
		}
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	if firstTrack == nil {
		return nil, nil, nil
	}
	if secondTrack == nil {
		return firstTrack, firstTrack, nil
	}

	return firstTrack, secondTrack, nil
}

func (qs *QueueStore) SpinQueue() error {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	tx, err := qs.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var firstTrackID string
	var firstTrackQueueID int

	query := `SELECT id, track_id FROM queue ORDER BY id ASC LIMIT 1`
	err = tx.QueryRow(query).Scan(&firstTrackQueueID, &firstTrackID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil // Queue is empty
		}
		return fmt.Errorf("failed to get first track: %w", err)
	}

	var maxID int

	err = tx.QueryRow(`SELECT MAX(id) FROM queue`).Scan(&maxID)
	if err != nil {
		return fmt.Errorf("failed to get max ID: %w", err)
	}

	query = `UPDATE queue SET id = ? WHERE id = ?`
	_, err = tx.Exec(query, maxID+1, firstTrackQueueID)
	if err != nil {
		return fmt.Errorf("failed to update first track ID: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
