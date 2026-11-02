package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/cheatsnake/airstation/internal/voiceover"
)

const voiceOverColumns = `
	v.id, v.voice_track_id, COALESCE(t.name, ''), v.target,
	COALESCE(v.target_track_id, ''), v.fade_in, v.fade_out, v.duck_db,
	v.start_offset, v.status, COALESCE(v.applied_track_id, ''),
	COALESCE(v.mix_path, ''), COALESCE(v.mix_id, ''), v.created_at`

type VoiceOverStore struct {
	db    *sql.DB
	mutex *sync.Mutex
}

func NewVoiceOverStore(db *sql.DB, mutex *sync.Mutex) VoiceOverStore {
	return VoiceOverStore{
		db:    db,
		mutex: mutex,
	}
}

func scanVoiceOver(row rowScanner) (*voiceover.VoiceOver, error) {
	vo := &voiceover.VoiceOver{}
	err := row.Scan(
		&vo.ID, &vo.VoiceTrackID, &vo.VoiceName, &vo.Target,
		&vo.TargetTrackID, &vo.FadeIn, &vo.FadeOut, &vo.DuckDB,
		&vo.StartOffset, &vo.Status, &vo.AppliedTrackID,
		&vo.MixPath, &vo.MixID, &vo.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return vo, nil
}

func (vs *VoiceOverStore) Add(vo *voiceover.VoiceOver) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	query := `
		INSERT INTO voice_overs
			(id, voice_track_id, target, target_track_id, fade_in, fade_out, duck_db,
			 start_offset, status, applied_track_id, mix_path, mix_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var targetTrackID any
	if vo.TargetTrackID != "" {
		targetTrackID = vo.TargetTrackID
	}

	_, err := vs.db.Exec(
		query,
		vo.ID, vo.VoiceTrackID, vo.Target, targetTrackID, vo.FadeIn, vo.FadeOut, vo.DuckDB,
		vo.StartOffset, vo.Status, vo.AppliedTrackID, vo.MixPath, vo.MixID, vo.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert voice-over: %w", err)
	}

	return nil
}

func (vs *VoiceOverStore) List() ([]*voiceover.VoiceOver, error) {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	query := `SELECT ` + voiceOverColumns + `
		FROM voice_overs v
		LEFT JOIN tracks t ON t.id = v.voice_track_id
		ORDER BY v.created_at DESC`

	rows, err := vs.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*voiceover.VoiceOver, 0)
	for rows.Next() {
		vo, err := scanVoiceOver(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, vo)
	}

	return list, rows.Err()
}

func (vs *VoiceOverStore) ByID(id string) (*voiceover.VoiceOver, error) {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	query := `SELECT ` + voiceOverColumns + `
		FROM voice_overs v
		LEFT JOIN tracks t ON t.id = v.voice_track_id
		WHERE v.id = ?`

	vo, err := scanVoiceOver(vs.db.QueryRow(query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("voice-over not found")
		}
		return nil, err
	}

	return vo, nil
}

func (vs *VoiceOverStore) PendingFor(trackID, target string) (*voiceover.VoiceOver, error) {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	query := `SELECT ` + voiceOverColumns + `
		FROM voice_overs v
		LEFT JOIN tracks t ON t.id = v.voice_track_id
		WHERE v.status = 'pending' AND v.target = ?`

	args := []any{target}
	if target == voiceover.TargetCurrent {
		query += " AND v.target_track_id = ?"
		args = append(args, trackID)
	} else {
		query += " AND (v.target_track_id IS NULL OR v.target_track_id = '' OR v.target_track_id = ?)"
		args = append(args, trackID)
	}
	query += " ORDER BY v.created_at ASC LIMIT 1"

	vo, err := scanVoiceOver(vs.db.QueryRow(query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return vo, nil
}

func (vs *VoiceOverStore) RenderedFor(trackID string) (*voiceover.VoiceOver, error) {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	query := `SELECT ` + voiceOverColumns + `
		FROM voice_overs v
		LEFT JOIN tracks t ON t.id = v.voice_track_id
		WHERE v.status = 'rendered' AND v.applied_track_id = ?
		ORDER BY v.created_at DESC LIMIT 1`

	vo, err := scanVoiceOver(vs.db.QueryRow(query, trackID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return vo, nil
}

func (vs *VoiceOverStore) MarkRendered(id, appliedTrackID, mixPath, mixID string) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	_, err := vs.db.Exec(
		`UPDATE voice_overs
		 SET status = 'rendered', applied_track_id = ?, mix_path = ?, mix_id = ?
		 WHERE id = ?`,
		appliedTrackID, mixPath, mixID, id,
	)
	if err != nil {
		return fmt.Errorf("failed to mark voice-over as rendered: %w", err)
	}

	return nil
}

func (vs *VoiceOverStore) MarkStatus(id, status string) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	_, err := vs.db.Exec(`UPDATE voice_overs SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return fmt.Errorf("failed to update voice-over status: %w", err)
	}

	return nil
}
