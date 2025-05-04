package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	sqltool "github.com/cheatsnake/airstation/internal/tools/sql"
	"github.com/cheatsnake/airstation/internal/tools/ulid"
	"github.com/cheatsnake/airstation/internal/track"

	_ "modernc.org/sqlite"
)

type TrackStore struct {
	db    *sql.DB
	log   *slog.Logger
	mutex sync.Mutex
}

func Open(dbPath string, log *slog.Logger) (*TrackStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(1)
	log.Info("Sqlite database connected.")

	tracksTable := `
	CREATE TABLE IF NOT EXISTS tracks (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		path TEXT NOT NULL,
		duration REAL NOT NULL,
		bitRate INTEGER NOT NULL
	);`
	_, err = db.Exec(tracksTable)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracks table: %w", err)
	}

	indexQuery := `CREATE INDEX IF NOT EXISTS idx_tracks_name ON tracks (name COLLATE NOCASE);`
	_, err = db.Exec(indexQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to create index on tracks.name: %w", err)
	}

	queueTable := `
		CREATE TABLE IF NOT EXISTS queue (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			track_id TEXT NOT NULL UNIQUE,
			FOREIGN KEY (track_id) REFERENCES tracks (id)
		);`
	_, err = db.Exec(queueTable)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue table: %w", err)
	}

	playbackHistoryTable := `
		CREATE TABLE IF NOT EXISTS playback_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			played_at INTEGER NOT NULL,
			track_name TEXT NOT NULL
		);`
	_, err = db.Exec(playbackHistoryTable)
	if err != nil {
		return nil, fmt.Errorf("failed to create table for playback history: %w", err)
	}

	playedAtIndexQuery := `CREATE INDEX IF NOT EXISTS idx_playback_history_played_at ON playback_history(played_at);`
	_, err = db.Exec(playedAtIndexQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to create index on playback_history.played_at: %w", err)
	}

	return &TrackStore{db: db}, nil
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

func (ts *TrackStore) Queue() ([]*track.Track, error) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	tracks := make([]*track.Track, 0, 10)

	query := `
		SELECT t.id, t.name, t.path, t.duration, t.bitRate
		FROM tracks t
		JOIN queue q ON t.id = q.track_id
		ORDER BY q.id ASC`
	rows, err := ts.db.Query(query)
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

func (ts *TrackStore) AddToQueue(tracks []*track.Track) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	query := `
			INSERT INTO queue (track_id)
			VALUES (?)
			ON CONFLICT (track_id) DO NOTHING
		`

	for _, track := range tracks {
		_, err := ts.db.Exec(query, track.ID)
		if err != nil {
			return fmt.Errorf("failed to add track to queue: %w", err)
		}
	}

	return nil
}

func (ts *TrackStore) RemoveFromQueue(trackIDs []string) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	query := `DELETE FROM queue WHERE track_id = ?`
	for _, id := range trackIDs {
		_, err := ts.db.Exec(query, id)
		if err != nil {
			return fmt.Errorf("failed to remove track from queue: %w", err)
		}
	}

	return nil
}

func (ts *TrackStore) ReorderQueue(trackIDs []string) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	_, err := ts.db.Exec(`DELETE FROM queue`)
	if err != nil {
		return fmt.Errorf("failed to clear queue: %w", err)
	}

	query := `INSERT INTO queue (track_id) VALUES (?)`
	for _, id := range trackIDs {
		_, err := ts.db.Exec(query, id)
		if err != nil {
			return fmt.Errorf("failed to reorder queue: %w", err)
		}
	}

	return nil
}

func (ts *TrackStore) CurrentAndNextTrack() (*track.Track, *track.Track, error) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	query := `
	SELECT t.id, t.name, t.path, t.duration, t.bitRate
	FROM tracks t
	JOIN queue q ON t.id = q.track_id
	ORDER BY q.id ASC
	LIMIT 2`
	rows, err := ts.db.Query(query)
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

func (ts *TrackStore) SpinQueue() error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	tx, err := ts.db.Begin()
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

func (ts *TrackStore) AddPlaybackHistory(playedAt int64, trackName string) error {
	query := `INSERT INTO playback_history (played_at, track_name) VALUES (?, ?)`

	_, err := ts.db.Exec(query, playedAt, trackName)
	if err != nil {
		return fmt.Errorf("failed to insert playback entry: %v", err)
	}

	return nil
}

func (ts *TrackStore) RecentPlaybackHistory() ([]*track.PlaybackHistory, error) {
	query := `
		SELECT id, played_at, track_name 
		FROM playback_history 
		WHERE played_at >= (strftime('%s', 'now') - 24 * 60 * 60)
		ORDER BY played_at DESC`

	rows, err := ts.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*track.PlaybackHistory
	for rows.Next() {
		var item track.PlaybackHistory
		if err := rows.Scan(&item.ID, &item.PlayedAt, &item.TrackName); err != nil {
			return nil, err
		}
		history = append(history, &item)
	}
	return history, nil
}

func (ts *TrackStore) DeleteOldPlaybackHistory() (int64, error) {
	query := `
		DELETE FROM playback_history 
		WHERE played_at < (strftime('%s', 'now') - 30 * 24 * 60 * 60)`

	result, err := ts.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old entries: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %v", err)
	}

	return rowsAffected, nil
}

func (ts *TrackStore) Close() error {
	return ts.db.Close()
}
