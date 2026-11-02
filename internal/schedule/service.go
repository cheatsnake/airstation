package schedule

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/cheatsnake/airstation/internal/pkg/ulid"
	"github.com/cheatsnake/airstation/internal/track"
)

const (
	stateLastInterludeAt = "last_interlude_at"
	stateLastTopHour     = "last_top_hour"
)

type Service struct {
	store   Store
	tracks  TrackStore
	history HistoryStore
	log     *slog.Logger
}

func NewService(store Store, tracks TrackStore, history HistoryStore, log *slog.Logger) *Service {
	return &Service{
		store:   store,
		tracks:  tracks,
		history: history,
		log:     log,
	}
}

func (s *Service) Rules() ([]*Rule, error) {
	return s.store.Rules()
}

func (s *Service) AddRule(name, kind string, everyNTracks, atMinute int, enabled bool) (*Rule, error) {
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return nil, fmt.Errorf("rule name must be at least 2 characters")
	}
	if !track.IsInterlude(kind) {
		return nil, fmt.Errorf("rule kind must be jingle, ad or station_id")
	}
	if err := validateTriggers(everyNTracks, atMinute); err != nil {
		return nil, err
	}

	rule := &Rule{
		ID:           ulid.New(),
		Name:         name,
		Kind:         kind,
		EveryNTracks: everyNTracks,
		AtMinute:     atMinute,
		Enabled:      enabled,
		CreatedAt:    time.Now().Unix(),
	}

	if err := s.store.AddRule(rule); err != nil {
		return nil, err
	}

	return rule, nil
}

func (s *Service) EditRule(id, name, kind string, everyNTracks, atMinute int, enabled bool) (*Rule, error) {
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return nil, fmt.Errorf("rule name must be at least 2 characters")
	}
	if !track.IsInterlude(kind) {
		return nil, fmt.Errorf("rule kind must be jingle, ad or station_id")
	}
	if err := validateTriggers(everyNTracks, atMinute); err != nil {
		return nil, err
	}

	rule := &Rule{
		ID:           id,
		Name:         name,
		Kind:         kind,
		EveryNTracks: everyNTracks,
		AtMinute:     atMinute,
		Enabled:      enabled,
		CreatedAt:    time.Now().Unix(),
	}
	if err := s.store.EditRule(rule); err != nil {
		return nil, err
	}

	return rule, nil
}

func (s *Service) DeleteRule(id string) error {
	return s.store.DeleteRule(id)
}

// Tick returns an interlude that is due right now, if any. queue is the
// current play order (the first item is what is playing).
func (s *Service) Tick(now time.Time, queue []*track.Track, isPlaying bool) (*Decision, error) {
	if !isPlaying || len(queue) == 0 {
		return nil, nil
	}

	// Do not stack interludes: if one is already playing or queued next,
	// wait until it has passed.
	for i, t := range queue {
		if i >= 2 {
			break
		}
		if t.Kind != track.KindMusic {
			return nil, nil
		}
	}

	rules, err := s.store.Rules()
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, nil
	}

	lastInterludeAt := s.stateInt(stateLastInterludeAt)
	if lastInterludeAt == 0 {
		// Start the counter now so a fresh install does not immediately
		// fire an interlude based on historical playback.
		_ = s.store.SetState(stateLastInterludeAt, strconv.FormatInt(now.Unix(), 10))
		return nil, nil
	}

	lastTopHour := s.stateInt(stateLastTopHour)
	playedMusic, err := s.history.MusicPlayedSince(lastInterludeAt)
	if err != nil {
		return nil, err
	}

	// Clock triggers (top of the hour, half hour, etc.) take priority.
	for _, rule := range rules {
		if !clockDue(rule, now, lastTopHour) {
			continue
		}
		if decision := s.pick(rule, queue); decision != nil {
			return decision, nil
		}
	}

	// Rotation triggers: every N music tracks.
	for _, rule := range rules {
		if !rotationDue(rule, playedMusic) {
			continue
		}
		if decision := s.pick(rule, queue); decision != nil {
			return decision, nil
		}
	}

	return nil, nil
}

func clockDue(rule *Rule, now time.Time, lastTopHour int64) bool {
	if !rule.Enabled || rule.AtMinute < 0 {
		return false
	}
	if now.Minute() != rule.AtMinute {
		return false
	}
	return lastTopHour < hourStart(now)
}

// hourStart returns the Unix timestamp of the start of the local wall-clock
// hour. Using the local hour (rather than truncating the absolute timestamp)
// keeps clock rules correct in timezones with half-hour offsets.
func hourStart(now time.Time) int64 {
	year, month, day := now.Date()
	return time.Date(year, month, day, now.Hour(), 0, 0, 0, now.Location()).Unix()
}

func rotationDue(rule *Rule, playedMusic int) bool {
	if !rule.Enabled || rule.EveryNTracks <= 0 {
		return false
	}
	return playedMusic >= rule.EveryNTracks
}

func (s *Service) MarkPlayed(decision *Decision, now time.Time) error {
	if decision == nil {
		return nil
	}

	if err := s.store.SetState(stateLastInterludeAt, strconv.FormatInt(now.Unix(), 10)); err != nil {
		return err
	}
	if decision.Rule.AtMinute >= 0 {
		if err := s.store.SetState(stateLastTopHour, strconv.FormatInt(hourStart(now), 10)); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) pick(rule *Rule, queue []*track.Track) *Decision {
	exclude := make([]string, 0, len(queue))
	for _, t := range queue {
		exclude = append(exclude, t.ID)
	}

	selected, err := s.tracks.RandomByKind(rule.Kind, exclude)
	if err != nil {
		s.log.Warn("Failed to pick interlude: " + err.Error())
		return nil
	}
	if selected == nil {
		return nil
	}

	return &Decision{Rule: rule, Track: selected}
}

func (s *Service) stateInt(key string) int64 {
	raw, err := s.store.GetState(key)
	if err != nil || raw == "" {
		return 0
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0
	}
	return value
}

func validateTriggers(everyNTracks, atMinute int) error {
	if everyNTracks < 0 || everyNTracks > 1000 {
		return fmt.Errorf("every_n_tracks must be between 0 and 1000")
	}
	if atMinute < -1 || atMinute > 59 {
		return fmt.Errorf("at_minute must be -1 or between 0 and 59")
	}
	if everyNTracks == 0 && atMinute < 0 {
		return fmt.Errorf("a rule needs either every_n_tracks or at_minute")
	}
	return nil
}
