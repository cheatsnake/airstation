package sqlite

import (
	"database/sql"
	"errors"
	"sync"

	"github.com/cheatsnake/airstation/internal/schedule"
)

type ScheduleStore struct {
	db    *sql.DB
	mutex *sync.Mutex
}

func NewScheduleStore(db *sql.DB, mutex *sync.Mutex) ScheduleStore {
	return ScheduleStore{
		db:    db,
		mutex: mutex,
	}
}

func (ss *ScheduleStore) Rules() ([]*schedule.Rule, error) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	rows, err := ss.db.Query(`
		SELECT id, name, kind, every_n_tracks, at_minute, enabled, created_at
		FROM schedule_rules
		ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]*schedule.Rule, 0)
	for rows.Next() {
		rule, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, rows.Err()
}

func (ss *ScheduleStore) AddRule(rule *schedule.Rule) error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	_, err := ss.db.Exec(`
		INSERT INTO schedule_rules (id, name, kind, every_n_tracks, at_minute, enabled, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rule.ID, rule.Name, rule.Kind, rule.EveryNTracks, rule.AtMinute, boolToInt(rule.Enabled), rule.CreatedAt,
	)
	return err
}

func (ss *ScheduleStore) EditRule(rule *schedule.Rule) error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	_, err := ss.db.Exec(`
		UPDATE schedule_rules
		SET name = ?, kind = ?, every_n_tracks = ?, at_minute = ?, enabled = ?
		WHERE id = ?`,
		rule.Name, rule.Kind, rule.EveryNTracks, rule.AtMinute, boolToInt(rule.Enabled), rule.ID,
	)
	return err
}

func (ss *ScheduleStore) DeleteRule(id string) error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	_, err := ss.db.Exec(`DELETE FROM schedule_rules WHERE id = ?`, id)
	return err
}

func (ss *ScheduleStore) GetState(key string) (string, error) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	var value string
	err := ss.db.QueryRow(`SELECT value FROM scheduler_state WHERE key = ?`, key).Scan(&value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}

	return value, nil
}

func (ss *ScheduleStore) SetState(key, value string) error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	_, err := ss.db.Exec(`
		INSERT INTO scheduler_state (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	)
	return err
}

func scanRule(row rowScanner) (*schedule.Rule, error) {
	rule := &schedule.Rule{}
	var enabled int

	err := row.Scan(
		&rule.ID, &rule.Name, &rule.Kind, &rule.EveryNTracks,
		&rule.AtMinute, &enabled, &rule.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	rule.Enabled = enabled == 1

	return rule, nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
