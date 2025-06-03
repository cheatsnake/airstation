// Package trackservice provides services related to audio track management.
package track

import (
	"fmt"
	"log/slog"
	"math"
	"path/filepath"
	"strings"

	"github.com/cheatsnake/airstation/internal/ffmpeg"
	"github.com/cheatsnake/airstation/internal/hls"
	"github.com/cheatsnake/airstation/internal/tools/fs"
)

// Service provides audio processing functionalities by interacting with a database and the FFmpeg CLI.
type Service struct {
	store     Store       // An instance of Storage for managing audio file storage.
	ffmpegCLI *ffmpeg.CLI // A pointer to the FFmpeg CLI wrapper for executing media processing commands.
	log       *slog.Logger

	LoadedTracksNotify chan int // Notification of the number of loaded tracks
}

// New creates and returns a new instance of Service.
//
// Parameters:
//   - store: An implementation of TrackStore for managing audio file storage.
//   - ffmpegCLI: A pointer to the FFmpeg CLI wrapper for executing media processing commands.
//
// Returns:
//   - A pointer to an initialized Service instance.
func NewService(store Store, ffmpegCLI *ffmpeg.CLI, log *slog.Logger) *Service {
	return &Service{
		store:     store,
		ffmpegCLI: ffmpegCLI,
		log:       log,

		LoadedTracksNotify: make(chan int),
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
func (s *Service) AddTrack(name, path string) (*Track, error) {
	metadata, err := s.ffmpegCLI.AudioMetadata(path)
	if err != nil {
		return nil, err
	}

	modDuration, err := s.modifyTrackDuration(path, metadata)
	if err != nil {
		return nil, err
	}

	if modDuration < minAllowedTrackDuration {
		return nil, fmt.Errorf("%s is too short for streaming", name)
	}

	if modDuration > maxAllowedTrackDuration {
		return nil, fmt.Errorf("%s is too large for streaming", name)
	}

	trackName := defineTrackName(name, metadata.Name)
	newTrack, err := s.store.AddTrack(trackName, path, modDuration, metadata.BitRate)
	if err != nil {
		return nil, err
	}

	return newTrack, nil
}

// PrepareTrack converts the audio file at filePath to AAC format with a fixed bitrate,
// saving the output to a new file with an .m4a extension.
//
// Parameters:
//   - filePath: The full path of the original audio file.
//
// Returns:
//   - The path to the converted .m4a file, or an error if the conversion fails.
func (s *Service) PrepareTrack(filePath string) (string, error) {
	newPath := replaceExtension(filePath, m4aExtension)
	err := s.ffmpegCLI.ConvertAudioToAAC(filePath, newPath, defaultAudioBitRate)
	if err != nil {
		return "", err
	}

	return newPath, nil
}

// Tracks retrieves a paginated list of tracks from the store, applying optional search, sort, and order.
//
// Parameters:
//   - page: The page number of results.
//   - limit: The number of results per page.
//   - search: A string to filter track names.
//   - sortBy: The field to sort by (id, name, or duration).
//   - sortOrder: The order of sorting (asc or desc).
//
// Returns:
//   - A TracksPage object with paginated track data, or an error.
func (s *Service) Tracks(page, limit int, search, sortBy, sortOrder string) (*Page, error) {
	if sortBy != "id" && sortBy != "name" && sortBy != "duration" {
		sortBy = "id"
	}

	if sortOrder != "asc" {
		sortOrder = "desc"
	}

	tracks, total, err := s.store.Tracks(page, limit, search, sortBy, sortOrder)
	if err != nil {
		return nil, err
	}

	return &Page{
		Tracks: tracks,
		Page:   page,
		Limit:  limit,
		Total:  total,
	}, nil
}

// DeleteTracks deletes tracks from the database and also removes their files from disk.
//
// Parameters:
//   - ids: A slice of strings contains track IDs.
//
// Returns:
//   - An error if deletion fails.
func (s *Service) DeleteTracks(ids []string) error {
	tracks, err := s.store.TracksByIDs(ids)
	if err != nil {
		return err
	}

	err = s.store.DeleteTracks(ids)
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

// FindTracks fetches track records by their IDs.
//
// Parameters:
//   - ids: A slice of strings contains track IDs.
//
// Returns:
//   - A slice of Track pointers or an error.
func (s *Service) FindTracks(ids []string) ([]*Track, error) {
	tracks, err := s.store.TracksByIDs(ids)
	return tracks, err
}

// MakeHLSPlaylist generates an HLS playlist for streaming using FFmpeg.
//
// Parameters:
//   - trackPath: The path of the audio track to segment.
//   - outDir: Output directory for the HLS segments and playlist.
//   - segName: Prefix for the segment files.
//   - segDuration: Duration of each HLS segment in seconds.
//
// Returns:
//   - An error if playlist generation fails.
func (s *Service) MakeHLSPlaylist(trackPath string, outDir string, segName string, segDuration int) error {
	err := s.ffmpegCLI.MakeHLSPlaylist(trackPath, outDir, segName, segDuration)
	return err
}

// LoadTracksFromDisk scans a directory for audio files, converts them if needed,
// adds them to the store, and deletes the original copies.
//
// Parameters:
//   - tracksDir: Directory path to load tracks from.
//
// Returns:
//   - A slice of loaded Track pointers, or an error.
func (s *Service) LoadTracksFromDisk(tracksDir string) ([]*Track, error) {
	tracks := make([]*Track, 0)

	mp3Filenames, err := fs.ListFilesFromDir(tracksDir, mp3Extension)
	if err != nil {
		return tracks, err
	}

	aacFilenames, err := fs.ListFilesFromDir(tracksDir, aacExtension)
	if err != nil {
		return tracks, err
	}

	wavFilenames, err := fs.ListFilesFromDir(tracksDir, wavExtension)
	if err != nil {
		return tracks, err
	}

	trackFilenames := make([]string, 0, len(mp3Filenames)+len(aacFilenames)+len(wavFilenames))
	trackFilenames = append(trackFilenames, mp3Filenames...)
	trackFilenames = append(trackFilenames, aacFilenames...)
	trackFilenames = append(trackFilenames, wavFilenames...)

	for _, trackFilename := range trackFilenames {
		trackPath := filepath.Join(tracksDir, trackFilename)
		preparedTrackPath, err := s.PrepareTrack(trackPath)
		if err != nil {
			s.log.Warn("Failed to prepare a track for streaming: " + err.Error())
			return tracks, err
		}

		track, err := s.AddTrack(trackFilename, preparedTrackPath)
		if err != nil {
			s.log.Warn("Failed to save track to database: " + err.Error())
			return tracks, err
		}

		err = fs.DeleteFile(trackPath)
		if err != nil {
			s.log.Warn("Failed to delete original copy of prepared track: " + err.Error())
		}

		tracks = append(tracks, track)
	}

	if len(tracks) > 0 {
		s.log.Info(fmt.Sprintf("Loaded %d new track(s) from disk.", len(tracks)))
		s.LoadedTracksNotify <- len(tracks)
	}

	return tracks, nil
}

// modifyTrackDuration changes the original track duration (slightly) to avoid small HLS segments.
func (s *Service) modifyTrackDuration(path string, metadata ffmpeg.AudioMetadata) (float64, error) {
	roundDur := roundDuration(metadata.Duration, hls.DefaultMaxSegmentDuration)
	roundDur -= 0.001 // need to avoid extra ms after padding/trimming

	if err := s.ffmpegCLI.TrimAudio(path, roundDur); err != nil {
		return 0, err
	}

	return roundDur, nil
}

// roundDuration define proper track length to be multiple for segment duration.
func roundDuration(trackDuration, segmentDuration float64) float64 {
	remainder := math.Mod(trackDuration, segmentDuration)

	// if the difference is not significant (less than 1.2 second), just crop it
	if remainder < 1.2 {
		return math.Floor(trackDuration - remainder)
	}

	// padding := segmentDuration - remainder
	// return math.Floor(trackDuration + padding)
	return math.Floor(trackDuration)
}

func defineTrackName(fileName, metaName string) string {
	if len(metaName) != 0 {
		return metaName
	}

	name := strings.ReplaceAll(fileName, ".mp3", "")
	name = strings.ReplaceAll(name, ".aac", "")
	name = strings.ReplaceAll(name, ".wav", "")
	name = strings.ReplaceAll(name, "_", " ")

	return name
}

func replaceExtension(path string, newExt string) string {
	if newExt != "" && !strings.HasPrefix(newExt, ".") {
		newExt = "." + newExt
	}

	ext := filepath.Ext(path)
	name := path[:len(path)-len(ext)]

	return name + newExt
}
