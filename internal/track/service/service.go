// Package trackservice provides services related to audio track management.
package trackservice

import (
	"math"

	"github.com/cheatsnake/airstation/internal/ffmpeg"
	"github.com/cheatsnake/airstation/internal/hls"
	"github.com/cheatsnake/airstation/internal/storage"
	"github.com/cheatsnake/airstation/internal/track"
)

// Service provides audio processing functionalities by interacting with a database and the FFmpeg CLI.
type Service struct {
	store     storage.TrackStore // An instance of TrackStore for managing audio file storage.
	ffmpegCLI *ffmpeg.CLI        // A pointer to the FFmpeg CLI wrapper for executing media processing commands.
}

// New creates and returns a new instance of Service.
//
// Parameters:
//   - store: An implementation of TrackStore for managing audio file storage.
//   - ffmpegCLI: A pointer to the FFmpeg CLI wrapper for executing media processing commands.
//
// Returns:
//   - A pointer to an initialized Service instance.
func New(store storage.TrackStore, ffmpegCLI *ffmpeg.CLI) *Service {
	return &Service{
		store:     store,
		ffmpegCLI: ffmpegCLI,
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

	newTrack, err := s.store.AddTrack(name, path, modDuration, metadata.BitRate)
	if err != nil {
		return nil, err
	}

	return newTrack, nil
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
