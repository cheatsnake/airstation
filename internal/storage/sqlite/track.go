package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	sqltool "github.com/cheatsnake/airstation/internal/tools/sql"
	"github.com/cheatsnake/airstation/internal/tools/ulid"
	"github.com/cheatsnake/airstation/internal/track"
)

type TrackStore struct {
	db    *sql.DB
	mutex *sync.Mutex
}

func NewTrackStore(db *sql.DB, mutex *sync.Mutex) TrackStore {
	return TrackStore{
		db:    db,
		mutex: mutex,
	}
}

func (ts *TrackStore) Tracks(page, limit int, search, sortBy, sortOrder string) ([]*track.Track, int, error) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	countQuery := "SELECT COUNT(*) FROM tracks"
	if search != "" {
		countQuery += " WHERE LOWER(name) LIKE LOWER(?)"
	}

	var total int
	var err error
	tracks := make([]*track.Track, 0, limit)

	if search != "" {
		err = ts.db.QueryRow(countQuery, "%"+search+"%").Scan(&total)
	} else {
		err = ts.db.QueryRow(countQuery).Scan(&total)
	}
	if err != nil {
		return tracks, 0, fmt.Errorf("failed to get total track count: %w", err)
	}

	query := "SELECT id, name, path, duration, bitRate FROM tracks"
	if search != "" {
		query += " WHERE name LIKE ?"
	}
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT ? OFFSET ?", sortBy, sortOrder)

	var rows *sql.Rows
	offset := (page - 1) * limit
	if search != "" {
		rows, err = ts.db.Query(query, "%"+strings.ToLower(search)+"%", limit, offset)
	} else {
		rows, err = ts.db.Query(query, limit, offset)
	}
	if err != nil {
		return tracks, 0, fmt.Errorf("failed to query tracks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var track track.Track
		err := rows.Scan(&track.ID, &track.Name, &track.Path, &track.Duration, &track.BitRate)
		if err != nil {
			return tracks, 0, fmt.Errorf("failed to scan track: %w", err)
		}
		tracks = append(tracks, &track)
	}

	if err = rows.Err(); err != nil {
		return tracks, 0, fmt.Errorf("error iterating over rows: %w", err)
	}

	return tracks, total, nil
}

func (ts *TrackStore) AddTrack(name, path string, duration float64, bitRate int) (*track.Track, error) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	id := ulid.New()
	track := &track.Track{
		ID:       id,
		Name:     name,
		Path:     path,
		Duration: duration,
		BitRate:  bitRate,
	}

	query := `INSERT INTO tracks (id, name, path, duration, bitRate) VALUES (?, ?, ?, ?, ?)`
	_, err := ts.db.Exec(query, track.ID, track.Name, track.Path, track.Duration, track.BitRate)
	if err != nil {
		return nil, fmt.Errorf("failed to insert track: %w", err)
	}

	return track, nil
}

func (ts *TrackStore) DeleteTracks(IDs []string) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	query := `DELETE FROM tracks WHERE id = ?`
	for _, id := range IDs {
		_, err := ts.db.Exec(query, id)
		if err != nil {
			return fmt.Errorf("failed to delete track with ID %s: %w", id, err)
		}
	}

	return nil
}

func (ts *TrackStore) EditTrack(track *track.Track) (*track.Track, error) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	query := `
	UPDATE tracks
	SET name = ?,
		path = ?,
		duration = ?,
		bitRate = ?
	WHERE id = ?`
	_, err := ts.db.Exec(query, track.Name, track.Path, track.Duration, track.BitRate, track.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update track: %w", err)
	}

	return track, nil
}

func (ts *TrackStore) TrackByID(ID string) (*track.Track, error) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	query := `SELECT id, name, path, duration, bitRate FROM tracks WHERE id = ?`
	row := ts.db.QueryRow(query, ID)

	var track track.Track
	err := row.Scan(&track.ID, &track.Name, &track.Path, &track.Duration, &track.BitRate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("track with ID %s not found", ID)
		}
		return nil, fmt.Errorf("failed to scan track: %w", err)
	}

	return &track, nil
}

func (ts *TrackStore) TracksByIDs(IDs []string) ([]*track.Track, error) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	tracks := make([]*track.Track, 0, len(IDs))

	whereClause := sqltool.BuildInClause("id", len(IDs))
	query := fmt.Sprintf("SELECT id, name, path, duration, bitRate FROM tracks WHERE %s", whereClause)
	args := make([]interface{}, len(IDs))
	for i, id := range IDs {
		args[i] = id
	}

	rows, err := ts.db.Query(query, args...)
	if err != nil {
		return tracks, fmt.Errorf("failed to query tracks: %w", err)
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
