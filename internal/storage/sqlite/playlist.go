package sqlite

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/cheatsnake/airstation/internal/playlist"
	"github.com/cheatsnake/airstation/internal/tools/ulid"
	"github.com/cheatsnake/airstation/internal/track"
)

type PlaylistStore struct {
	db    *sql.DB
	mutex *sync.Mutex
}

func NewPlaylistStore(db *sql.DB, mutex *sync.Mutex) PlaylistStore {
	return PlaylistStore{
		db:    db,
		mutex: mutex,
	}
}

// AddPlaylist inserts a new playlist and associates tracks
func (ps *PlaylistStore) AddPlaylist(name, description string, trackIDs []string) (*playlist.Playlist, error) {
	id := ulid.New()

	tx, err := ps.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	_, err = tx.Exec(`INSERT INTO playlist (id, name, description) VALUES (?, ?, ?)`, id, name, description)
	if err != nil {
		return nil, err
	}

	for _, trackID := range trackIDs {
		_, err = tx.Exec(`INSERT OR IGNORE INTO playlist_track (playlist_id, track_id) VALUES (?, ?)`, id, trackID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return ps.Playlist(id)
}

// Playlists returns all playlists without tracks
func (ps *PlaylistStore) Playlists() ([]*playlist.Playlist, error) {
	query := `
		SELECT p.id, p.name, p.description, COUNT(pt.track_id) as track_count
		FROM playlist p
		LEFT JOIN playlist_track pt ON p.id = pt.playlist_id
		GROUP BY p.id
	`

	rows, err := ps.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	playlists := make([]*playlist.Playlist, 0)

	for rows.Next() {
		var p playlist.Playlist
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.TrackCount); err != nil {
			return nil, err
		}

		p.Tracks = []*track.Track{}
		playlists = append(playlists, &p)
	}

	return playlists, nil
}

// Playlist returns a playlist with all its tracks
func (ps *PlaylistStore) Playlist(id string) (*playlist.Playlist, error) {
	p := playlist.Playlist{Tracks: make([]*track.Track, 0)}

	err := ps.db.QueryRow(`SELECT id, name, description FROM playlist WHERE id = ?`, id).
		Scan(&p.ID, &p.Name, &p.Description)
	if err != nil {
		return nil, err
	}

	rows, err := ps.db.Query(`
		SELECT t.id, t.name, t.path, t.bitRate, t.duration
		FROM playlist_track pt
		JOIN tracks t ON pt.track_id = t.id
		WHERE pt.playlist_id = ?
		ORDER BY pt.position
	`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query playlist tracks: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var t track.Track
		if err := rows.Scan(&t.ID, &t.Name, &t.Path, &t.BitRate, &t.Duration); err != nil {
			return nil, fmt.Errorf("failed to scan track: %w", err)
		}
		p.Tracks = append(p.Tracks, &t)
	}

	p.TrackCount = len(p.Tracks)

	return &p, nil
}

// IsPlaylistExists checks if playlist with provided name exists
func (ps *PlaylistStore) IsPlaylistExists(name string) (bool, error) {
	var exists bool

	err := ps.db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM playlist WHERE name = ?
        )
    `, name).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

// EditPlaylist updates playlist and its tracks
func (ps *PlaylistStore) EditPlaylist(id, name, description string, trackIDs []string) error {
	tx, err := ps.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE playlist SET name = ?, description = ? WHERE id = ?`, name, description, id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM playlist_track WHERE playlist_id = ?`, id)
	if err != nil {
		return err
	}

	for position, trackID := range trackIDs {
		_, err = tx.Exec(
			`INSERT OR IGNORE INTO playlist_track (playlist_id, track_id, position) VALUES (?, ?, ?)`,
			id, trackID, position,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// DeletePlaylist deletes playlist and its track associations
func (ps *PlaylistStore) DeletePlaylist(id string) error {
	tx, err := ps.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM playlist_track WHERE playlist_id = ?`, id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM playlist WHERE id = ?`, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}
