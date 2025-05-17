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
func (s *Service) AddPlaybackHistory(trackName string) {
	err := s.store.AddPlaybackHistory(time.Now().Unix(), trackName)
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

// DeleteOldPlaybackHistory removes outdated playback history entries from the store.
func (s *Service) DeleteOldPlaybackHistory() {
	_, err := s.store.DeleteOldPlaybackHistory()
	if err != nil {
		s.log.Warn("Failed to delete old playback history: " + err.Error())
	}
}
