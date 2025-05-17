package queue

import (
	"path"
	"strings"
	"time"

	"github.com/cheatsnake/airstation/internal/hls"
	"github.com/cheatsnake/airstation/internal/tools/fs"
	"github.com/cheatsnake/airstation/internal/track"
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}

// Queue retrieves the current playback queue.
//
// Returns:
//   - A slice of Track pointers or an error.
func (s *Service) Queue() ([]*track.Track, error) {
	q, err := s.store.Queue()
	return q, err
}

// AddToQueue adds one or more tracks to the playback queue.
//
// Parameters:
//   - tracks: A slice of Track pointers to add.
//
// Returns:
//   - An error if the operation fails.
func (s *Service) AddToQueue(tracks []*track.Track) error {
	err := s.store.AddToQueue(tracks)
	return err
}

// ReorderQueue updates the order of tracks in the playback queue.
//
// Parameters:
//   - ids: A slice of strings contains track IDs.
//
// Returns:
//   - An error if reordering fails.
func (s *Service) ReorderQueue(ids []string) error {
	err := s.store.ReorderQueue(ids)
	return err
}

// RemoveFromQueue removes specific tracks from the playback queue.
//
// Parameters:
//   - ids: A slice of strings contains track IDs.
//
// Returns:
//   - An error if removal fails.
func (s *Service) RemoveFromQueue(ids []string) error {
	err := s.store.RemoveFromQueue(ids)
	return err
}

// SpinQueue rotates the playback queue, moving the current track to the end.
//
// Returns:
//   - An error if the operation fails.
func (s *Service) SpinQueue() error {
	err := s.store.SpinQueue()
	return err
}

// CurrentAndNextTrack retrieves the currently playing track and the next track in the queue.
//
// Returns:
//   - Pointers to the current and next tracks, and an error if retrieval fails.
func (s *Service) CurrentAndNextTrack() (*track.Track, *track.Track, error) {
	current, next, err := s.store.CurrentAndNextTrack()
	return current, next, err
}

// CleanupHLSPlaylists removes old HLS playlist files that are no longer needed.
//
// Parameters:
//   - dirPath: Directory containing the HLS playlist files.
//
// Returns:
//   - An error if file cleanup fails.
func (s *Service) CleanupHLSPlaylists(dirPath string) error {
	// waiting for all the listeners to listen to the last segments of ended track
	time.Sleep(hls.DefaultMaxSegmentDuration * 2 * time.Second)
	current, next, err := s.store.CurrentAndNextTrack()
	if err != nil {
		return err
	}

	utilized := []string{current.ID, next.ID}
	tmpFiles, err := fs.ListFilesFromDir(dirPath, "")
	if err != nil {
		return err
	}

	for _, tmpFile := range tmpFiles {
		keep := false
		for _, prefix := range utilized {
			if strings.HasPrefix(tmpFile, prefix) {
				keep = true
				break
			}
		}
		if !keep {
			fs.DeleteFile(path.Join(dirPath, tmpFile))
		}
	}

	return nil
}
