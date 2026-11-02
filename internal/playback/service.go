package playback

import (
	"log/slog"
	"time"
)

type Service struct {
	store Store
	log   *slog.Logger
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}

// AddPlaybackHistory logs a playback event for a given track.
//
// Parameters:
//   - trackName: The name of the track that was played.
//   - kind: The asset kind (music, jingle, ad, station_id or voice).
func (s *Service) AddPlaybackHistory(trackName, kind string) {
	err := s.store.AddPlaybackHistory(time.Now().Unix(), trackName, kind)
	if err != nil {
		s.log.Error("Failed to add playback history: " + err.Error())
	}
}

// RecentPlaybackHistory retrieves the most recent playback history records.
//
// Parameters:
//   - limit: The maximum number of history entries to retrieve.
//
// Returns:
//   - A slice of PlaybackHistory pointers, or an error.
func (s *Service) RecentPlaybackHistory(limit int) ([]*History, error) {
	history, err := s.store.RecentPlaybackHistory(limit)
	return history, err
}

// MusicPlayedSince returns how many music tracks have been played since ts.
// The scheduler uses it to implement "every N songs" interlude rules.
func (s *Service) MusicPlayedSince(ts int64) (int, error) {
	return s.store.MusicPlayedSince(ts)
}

// DeleteOldPlaybackHistory removes outdated playback history entries from the store.
func (s *Service) DeleteOldPlaybackHistory() {
	_, err := s.store.DeleteOldPlaybackHistory()
	if err != nil {
		s.log.Warn("Failed to delete old playback history: " + err.Error())
	}
}
