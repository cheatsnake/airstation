// Package schedule implements simple radio clock rules for automatically
// inserting jingles, ads and station IDs into the queue.
package schedule

import "github.com/cheatsnake/airstation/internal/track"

// Rule describes when an interlude should be inserted.
type Rule struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Kind         string `json:"kind"`
	EveryNTracks int    `json:"everyNTracks"` // 0 disables the "every N songs" trigger.
	AtMinute     int    `json:"atMinute"`     // -1 disables the clock trigger, 0..59 is the minute of the hour.
	Enabled      bool   `json:"enabled"`
	CreatedAt    int64  `json:"createdAt"`
}

// Decision is a due rule together with the asset selected for it.
type Decision struct {
	Rule  *Rule
	Track *track.Track
}

type Store interface {
	Rules() ([]*Rule, error)
	AddRule(rule *Rule) error
	EditRule(rule *Rule) error
	DeleteRule(id string) error
	GetState(key string) (string, error)
	SetState(key, value string) error
}

type TrackStore interface {
	RandomByKind(kind string, excludeIDs []string) (*track.Track, error)
}

type HistoryStore interface {
	MusicPlayedSince(ts int64) (int, error)
}
