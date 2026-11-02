package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	sqltool "github.com/cheatsnake/airstation/internal/pkg/sql"
	"github.com/cheatsnake/airstation/internal/pkg/ulid"
	"github.com/cheatsnake/airstation/internal/track"
)

const trackColumns = "id, name, path, duration, bitRate, kind, fade_in, fade_out, duck_db"
const trackColumnsAliased = "t.id, t.name, t.path, t.duration, t.bitRate, t.kind, t.fade_in, t.fade_out, t.duck_db"

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTrack(row rowScanner) (*track.Track, error) {
	t := &track.Track{}
	err := row.Scan(&t.ID, &t.Name, &t.Path, &t.Duration, &t.BitRate, &t.Kind, &t.FadeIn, &t.FadeOut, &t.DuckDB)
	if err != nil {
		return nil, err
	}
	return t, nil
}

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

func (ts *TrackStore) Tracks(page, limit int, search, sortBy, sortOrder, kind string) ([]*track.Track, int, error) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	where := make([]string, 0, 2)
	args := make([]any, 0, 2)

	if search != "" {
		where = append(where, "LOWER(name) LIKE LOWER(?)")
		args = append(args, "%"+search+"%")
	}
	if kind != "" {
		where = append(where, "kind = ?")
		args = append(args, kind)
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	var total int
	if err := ts.db.QueryRow("SELECT COUNT(*) FROM tracks"+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to get total track count: %w", err)
	}

	query := "SELECT " + trackColumns + " FROM tracks" + whereClause
	query += fmt.Sprintf(" ORDER BY %s %s LIMIT ? OFFSET ?", sortBy, sortOrder)

	offset := (page - 1) * limit
	queryArgs := append(append([]any{}, args...), limit, offset)

	rows, err := ts.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query tracks: %w", err)
	}
	defer rows.Close()

	tracks := make([]*track.Track, 0, limit)
	for rows.Next() {
		t, err := scanTrack(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan track: %w", err)
		}
		tracks = append(tracks, t)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating over rows: %w", err)
	}

	return tracks, total, nil
}

func (ts *TrackStore) AddTrack(name, path string, duration float64, bitRate int, kind string, fadeIn, fadeOut, duckDB float64) (*track.Track, error) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	id := ulid.New()
	t := &track.Track{
		ID:       id,
		Name:     name,
		Path:     path,
		Duration: duration,
		BitRate:  bitRate,
		Kind:     kind,
		FadeIn:   fadeIn,
		FadeOut:  fadeOut,
		DuckDB:   duckDB,
	}

	query := `INSERT INTO tracks (id, name, path, duration, bitRate, kind, fade_in, fade_out, duck_db)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	err := ts.insertTrack(query, t)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
		// Cart and voice assets often share a title with a song. Keep the
		// original label readable while making the database key unique.
		t.Name = fmt.Sprintf("%s (%s)", name, t.ID[len(t.ID)-4:])
		err = ts.insertTrack(query, t)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to insert track: %w", err)
	}

	return t, nil
}

func (ts *TrackStore) insertTrack(query string, t *track.Track) error {
	_, err := ts.db.Exec(query, t.ID, t.Name, t.Path, t.Duration, t.BitRate, t.Kind, t.FadeIn, t.FadeOut, t.DuckDB)
	return err
}

func (ts *TrackStore) DeleteTracks(IDs []string) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	tx, err := ts.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Remove references first so the tracks can be deleted even when they are
	// still queued, in a playlist, or used as the source of a voice-over.
	referencingQueries := []string{
		`DELETE FROM queue WHERE track_id = ?`,
		`DELETE FROM playlist_track WHERE track_id = ?`,
		`DELETE FROM voice_overs WHERE voice_track_id = ?`,
	}
	for _, id := range IDs {
		for _, query := range referencingQueries {
			if _, err := tx.Exec(query, id); err != nil {
				return fmt.Errorf("failed to remove references to track with ID %s: %w", id, err)
			}
		}
	}

	query := `DELETE FROM tracks WHERE id = ?`
	for _, id := range IDs {
		if _, err := tx.Exec(query, id); err != nil {
			return fmt.Errorf("failed to delete track with ID %s: %w", id, err)
		}
	}

	return tx.Commit()
}

func (ts *TrackStore) EditTrack(t *track.Track) (*track.Track, error) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	query := `
	UPDATE tracks
	SET name = ?,
		path = ?,
		duration = ?,
		bitRate = ?,
		kind = ?,
		fade_in = ?,
		fade_out = ?,
		duck_db = ?
	WHERE id = ?`
	_, err := ts.db.Exec(query, t.Name, t.Path, t.Duration, t.BitRate, t.Kind, t.FadeIn, t.FadeOut, t.DuckDB, t.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update track: %w", err)
	}

	return t, nil
}

func (ts *TrackStore) TrackByID(ID string) (*track.Track, error) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	query := "SELECT " + trackColumns + " FROM tracks WHERE id = ?"
	t, err := scanTrack(ts.db.QueryRow(query, ID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("track with ID %s not found", ID)
		}
		return nil, fmt.Errorf("failed to scan track: %w", err)
	}

	return t, nil
}

func (ts *TrackStore) TracksByIDs(IDs []string) ([]*track.Track, error) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	tracks := make([]*track.Track, 0, len(IDs))
	if len(IDs) == 0 {
		return tracks, nil
	}

	whereClause := sqltool.BuildInClause("id", len(IDs))
	query := fmt.Sprintf("SELECT %s FROM tracks WHERE %s", trackColumns, whereClause)
	args := make([]any, len(IDs))
	for i, id := range IDs {
		args[i] = id
	}

	rows, err := ts.db.Query(query, args...)
	if err != nil {
		return tracks, fmt.Errorf("failed to query tracks: %w", err)
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

// RandomByKind returns a random asset of the requested kind, excluding the
// supplied IDs. It is used by the scheduler to pick jingles/ads/IDs.
func (ts *TrackStore) RandomByKind(kind string, excludeIDs []string) (*track.Track, error) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	query := "SELECT " + trackColumns + " FROM tracks WHERE kind = ?"
	args := []any{kind}

	if len(excludeIDs) > 0 {
		query += fmt.Sprintf(" AND id NOT IN (%s)", sqltool.BuildInClause("id", len(excludeIDs)))
		for _, id := range excludeIDs {
			args = append(args, id)
		}
	}

	query += " ORDER BY RANDOM() LIMIT 1"

	t, err := scanTrack(ts.db.QueryRow(query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to pick random %s: %w", kind, err)
	}

	return t, nil
}
