// Package trackservice provides services related to audio track management.
package trackservice

import (
	"log/slog"
	"math"
	"strings"

	"github.com/cheatsnake/airstation/internal/ffmpeg"
	"github.com/cheatsnake/airstation/internal/hls"
	"github.com/cheatsnake/airstation/internal/storage"
	"github.com/cheatsnake/airstation/internal/tools/fs"
	"github.com/cheatsnake/airstation/internal/track"
)

// Service provides audio processing functionalities by interacting with a database and the FFmpeg CLI.
type Service struct {
	store     storage.TrackStore // An instance of TrackStore for managing audio file storage.
	ffmpegCLI *ffmpeg.CLI        // A pointer to the FFmpeg CLI wrapper for executing media processing commands.
	log       *slog.Logger
}

// New creates and returns a new instance of Service.
//
// Parameters:
//   - store: An implementation of TrackStore for managing audio file storage.
//   - ffmpegCLI: A pointer to the FFmpeg CLI wrapper for executing media processing commands.
//
// Returns:
//   - A pointer to an initialized Service instance.
func New(store storage.TrackStore, ffmpegCLI *ffmpeg.CLI, log *slog.Logger) *Service {
	return &Service{
		store:     store,
		ffmpegCLI: ffmpegCLI,
		log:       log,
	}
}

// AddTrack adds a new audio track to the database, extracting metadata and modifying its duration if necessary.
//
// Parameters:
//   - name: The name to assign to the new track.
//   - path: The file path of the audio track to be added.
//
// Returns:
//   - A pointer to the newly added Track, or an error if any step in the process fails.
func (s *Service) AddTrack(name, path string) (*track.Track, error) {
	metadata, err := s.ffmpegCLI.AudioMetadata(path)
	if err != nil {
		return nil, err
	}

	modDuration, err := s.modifyTrackDuration(path, metadata)
	if err != nil {
		return nil, err
	}

	trackName := defineTrackName(name, metadata.Name)
	newTrack, err := s.store.AddTrack(trackName, path, modDuration, metadata.BitRate)
	if err != nil {
		return nil, err
	}

	return newTrack, nil
}

func (s *Service) Tracks(page, limit int, search string) (*TracksPage, error) {
	tracks, total, err := s.store.Tracks(page, limit, search)
	if err != nil {
		return nil, err
	}

	return &TracksPage{
		Tracks: tracks,
		Page:   page,
		Limit:  limit,
		Total:  total,
	}, nil
}

func (s *Service) DeleteTracks(ids *TrackIDs) error {
	tracks, err := s.store.TracksByIDs(ids.IDs)
	if err != nil {
		return err
	}

	err = s.store.DeleteTracks(ids.IDs)
	if err != nil {
		return err
	}

	for _, t := range tracks {
		err := fs.DeleteFile(t.Path)
		if err != nil {
			s.log.Warn("Failed to delete track from disk: " + err.Error())
		}
	}

	return err
}

func (s *Service) FindTracks(ids *TrackIDs) ([]*track.Track, error) {
	tracks, err := s.store.TracksByIDs(ids.IDs)
	return tracks, err
}

func (s *Service) Queue() ([]*track.Track, error) {
	q, err := s.store.Queue()
	return q, err
}

func (s *Service) AddToQueue(tracks []*track.Track) error {
	err := s.store.AddToQueue(tracks)
	return err
}

func (s *Service) ReorderQueue(ids *TrackIDs) error {
	err := s.store.ReorderQueue(ids.IDs)
	return err
}

func (s *Service) RemoveFromQueue(ids *TrackIDs) error {
	err := s.store.RemoveFromQueue(ids.IDs)
	return err
}

func (s *Service) SpinQueue() error {
	err := s.store.SpinQueue()
	return err
}

func (s *Service) CurrentAndNextTrack() (*track.Track, *track.Track, error) {
	current, next, err := s.store.CurrentAndNextTrack()
	return current, next, err
}

func (s *Service) MakeHLSPlaylist(trackPath string, outDir string, segName string, segDuration int) error {
	err := s.ffmpegCLI.MakeHLSPlaylist(trackPath, outDir, segName, segDuration)
	return err
}

// modifyTrackDuration changes the original track duration (slightly) to avoid small HLS segments.
func (s *Service) modifyTrackDuration(path string, metadata ffmpeg.AudioMetadata) (float64, error) {
	roundDur := roundDuration(metadata.Duration, hls.DefaultMaxSegmentDuration)
	roundDur -= 0.1 // need to avoid extra ms after padding/trimming

	if roundDur > metadata.Duration {
		if err := s.ffmpegCLI.PadAudio(path, roundDur-metadata.Duration, metadata); err != nil {
			return 0, err
		}
	}

	if roundDur < metadata.Duration {
		if err := s.ffmpegCLI.TrimAudio(path, roundDur); err != nil {
			return 0, err
		}
	}

	return roundDur, nil
}

// roundDuration define proper track length to be multiple for segment duration.
func roundDuration(trackDuration, segmentDuration float64) float64 {
	remainder := math.Mod(trackDuration, segmentDuration)

	// if the difference is not significant (less than second), just crop it
	if remainder < 1 {
		return math.Floor(trackDuration - remainder)
	}

	padding := segmentDuration - remainder
	return math.Floor(trackDuration + padding)
}

func defineTrackName(fileName, metaName string) string {
	if len(metaName) != 0 {
		return metaName
	}

	name := strings.ReplaceAll(fileName, ".mp3", "")
	name = strings.ReplaceAll(name, ".aac", "")
	name = strings.ReplaceAll(name, "_", " ")

	return name
}
