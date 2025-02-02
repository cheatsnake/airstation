// Package track defines the structure and functions for managing audio track entities.
package track

import "github.com/cheatsnake/airstation/internal/tools/ulid"

// Track represents an audio track with its associated metadata.
type Track struct {
	ID       string  `json:"id"`       // A unique identifier for the track, typically generated using ULID.
	Name     string  `json:"name"`     // The name of the audio track.
	Path     string  `json:"path"`     // The file path of the audio track.
	Duration float64 `json:"duration"` // The duration of the audio track in seconds.
	BitRate  int     `json:"bitRate"`  // The bit rate of the audio track in kilobits per second (kbps).
}

// New creates and returns a new Track instance with the provided name, path, duration, and bit rate.
// It also generates a unique ID for the track.
//
// Parameters:
//   - name: The name of the track.
//   - path: The file path of the track.
//   - duration: The duration of the track in seconds.
//   - bitRate: The bit rate of the track in kilobits per second (kbps).
//
// Returns:
//   - A pointer to the newly created Track instance.
func New(name, path string, duration float64, bitRate int) *Track {
	return &Track{
		ID:       ulid.New(),
		Name:     name,
		Path:     path,
		Duration: duration,
		BitRate:  bitRate,
	}
}
