package sqlite

import (
	"database/sql"
	"errors"
	"sync"

	"github.com/cheatsnake/airstation/internal/station"
)

type StationStore struct {
	db    *sql.DB
	mutex *sync.Mutex
}

func NewStationStore(db *sql.DB, mutex *sync.Mutex) StationStore {
	return StationStore{
		db:    db,
		mutex: mutex,
	}
}

func (ss *StationStore) StationProperties() ([]*station.Property, error) {
	query := `
		SELECT key, value
		FROM station_properties
		ORDER BY key
	`

	rows, err := ss.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var properties []*station.Property
	for rows.Next() {
		var prop station.Property
		if err := rows.Scan(&prop.Key, &prop.Value); err != nil {
			return nil, err
		}
		properties = append(properties, &prop)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return properties, nil
}

func (ss *StationStore) UpsertStationProperty(key, value string) (*station.Property, error) {
	if key == "" {
		return nil, errors.New("key cannot be empty")
	}

	query := `
		INSERT INTO station_properties (key, value)
		VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = strftime('%s', 'now')
	`

	_, err := ss.db.Exec(query, key, value)
	if err != nil {
		return nil, err
	}

	return &station.Property{Key: key, Value: value}, nil
}

func (ss *StationStore) DeleteStationProperty(key string) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}

	query := "DELETE FROM station_properties WHERE key = ?"

	result, err := ss.db.Exec(query, key)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("property not found")
	}

	return nil
}
