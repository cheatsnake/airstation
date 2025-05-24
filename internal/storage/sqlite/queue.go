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
		SELECT t.id, t.name, t.path, t.duration, t.bitRate
		FROM tracks t
		JOIN queue q ON t.id = q.track_id
		ORDER BY q.id ASC`
	rows, err := qs.db.Query(query)
	if err != nil {
		return tracks, fmt.Errorf("failed to query tracks in queue: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var track track.Track
		err := rows.Scan(&track.ID, &track.Name, &track.Path, &track.Duration, &track.BitRate)
		if err != nil {
			return tracks, fmt.Errorf("failed to scan track: %w", err)
		}
		tracks = append(tracks, &track)
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
	SELECT t.id, t.name, t.path, t.duration, t.bitRate
	FROM tracks t
	JOIN queue q ON t.id = q.track_id
	ORDER BY q.id ASC
	LIMIT 2`
	rows, err := qs.db.Query(query)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query first and second tracks: %w", err)
	}
	defer rows.Close()

	var firstTrack, secondTrack track.Track
	count := 0

	for rows.Next() {
		if count == 0 {
			err := rows.Scan(&firstTrack.ID, &firstTrack.Name, &firstTrack.Path, &firstTrack.Duration, &firstTrack.BitRate)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to scan first track: %w", err)
			}
		} else if count == 1 {
			err := rows.Scan(&secondTrack.ID, &secondTrack.Name, &secondTrack.Path, &secondTrack.Duration, &secondTrack.BitRate)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to scan second track: %w", err)
			}
		}
		count++
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	if count == 0 {
		return nil, nil, nil
	} else if count == 1 {
		return &firstTrack, &firstTrack, nil
	}

	return &firstTrack, &secondTrack, nil
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
