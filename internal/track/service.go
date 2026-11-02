// Package trackservice provides services related to audio track management.
package track

import (
	"fmt"
	"log/slog"
	"math"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cheatsnake/airstation/internal/pkg/ffmpeg"
	"github.com/cheatsnake/airstation/internal/pkg/fs"
	"github.com/cheatsnake/airstation/internal/pkg/hls"
	"github.com/cheatsnake/airstation/internal/station"
)

// Service provides audio processing functionalities by interacting with a database and the FFmpeg CLI.
type Service struct {
	store     Store
	station   *station.Service
	ffmpegCLI *ffmpeg.CLI
	log       *slog.Logger

	LoadedTracksNotify chan int // Notification of the number of loaded tracks
	loadMutex          sync.Mutex
}

// New creates and returns a new instance of Service.
//
// Parameters:
//   - store: An implementation of TrackStore for managing audio file storage.
//   - ffmpegCLI: A pointer to the FFmpeg CLI wrapper for executing media processing commands.
//
// Returns:
//   - A pointer to an initialized Service instance.
func NewService(store Store, stationSvc *station.Service, ffmpegCLI *ffmpeg.CLI, log *slog.Logger) *Service {
	return &Service{
		store:     store,
		station:   stationSvc,
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
//   - kind: The asset kind (music, jingle, ad, station_id or voice).
//   - fadeIn/fadeOut/duckDB: Default mix settings stored on voice assets.
//
// Returns:
//   - A pointer to the newly added Track, or an error if any step in the process fails.
//
// audioBitRate returns the operator-configured target bitrate for converted
// audio, falling back to the built-in default when unset or unavailable.
func (s *Service) audioBitRate() int {
	if s.station != nil {
		if info, err := s.station.Info(); err == nil && info.AudioBitRate > 0 {
			return info.AudioBitRate
		}
	}
	return defaultAudioBitRate
}

func (s *Service) AddTrack(name, path, kind string, fadeIn, fadeOut, duckDB float64) (*Track, error) {
	if kind == "" {
		kind = KindMusic
	}
	if !IsValidKind(kind) {
		return nil, fmt.Errorf("unsupported track kind: %s", kind)
	}

	metadata, err := s.ffmpegCLI.AudioMetadata(path)
	if err != nil {
		return nil, err
	}

	modDuration, err := s.modifyTrackDuration(path, metadata)
	if err != nil {
		return nil, err
	}

	minDuration := float64(minAllowedTrackDuration)
	if kind != KindMusic {
		minDuration = minAllowedAssetDuration
	}
	if modDuration < minDuration {
		return nil, fmt.Errorf("%s is too short for streaming", name)
	}

	if modDuration > maxAllowedTrackDuration {
		return nil, fmt.Errorf("%s is too long for streaming", name)
	}

	if kind == KindVoice {
		if fadeIn <= 0 {
			fadeIn = 1.5
		}
		if fadeOut <= 0 {
			fadeOut = 2.0
		}
		if duckDB == 0 {
			duckDB = -12
		}
	} else {
		fadeIn, fadeOut, duckDB = 0, 0, 0
	}

	trackName := defineTrackName(name, metadata.Name)
	newTrack, err := s.store.AddTrack(trackName, path, modDuration, metadata.BitRate, kind, fadeIn, fadeOut, duckDB)
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
func (s *Service) PrepareTrack(filePath string, bitRate int) (string, error) {
	if bitRate <= 0 {
		bitRate = s.audioBitRate()
	}
	newPath := replaceExtension(filePath, m4aExtension)
	err := s.ffmpegCLI.ConvertAudioToAAC(filePath, newPath, bitRate)
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
func (s *Service) Tracks(page, limit int, search, sortBy, sortOrder, kind string) (*Page, error) {
	if sortBy != "id" && sortBy != "name" && sortBy != "duration" {
		sortBy = "id"
	}

	if sortOrder != "asc" {
		sortOrder = "desc"
	}

	if kind == "all" {
		kind = ""
	}
	if kind != "" && !IsValidKind(kind) {
		return nil, fmt.Errorf("unsupported track kind: %s", kind)
	}

	tracks, total, err := s.store.Tracks(page, limit, search, sortBy, sortOrder, kind)
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

// Carts returns every non-music asset (jingles, ads, station IDs and voice
// clips) that can be fired from the broadcast console.
func (s *Service) Carts() ([]*Track, error) {
	page, err := s.Tracks(1, 1000, "", "name", "asc", "")
	if err != nil {
		return nil, err
	}

	carts := make([]*Track, 0, len(page.Tracks))
	for _, t := range page.Tracks {
		if t.Kind != KindMusic {
			carts = append(carts, t)
		}
	}

	return carts, nil
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
	extensions := []string{mp3Extension, aacExtension, wavExtension, flacExtension}
	paths := make([]string, 0)

	for _, ext := range extensions {
		filenames, err := fs.ListFilesFromDir(tracksDir, ext)
		if err != nil {
			return nil, err
		}
		for _, filename := range filenames {
			paths = append(paths, filepath.Join(tracksDir, filename))
		}
	}

	tracks, _ := s.loadPaths(paths, KindMusic)
	return tracks, nil
}

// LoadUploadedTracks processes a specific set of freshly uploaded files. It is
// used by the HTTP upload handler so the caller can choose the asset kind
// (music, jingle, ad, station ID or voice) without re-scanning the directory.
// It returns the assets that were stored plus one error per rejected file.
func (s *Service) LoadUploadedTracks(paths []string, kind string) ([]*Track, []error) {
	if kind == "" {
		kind = KindMusic
	}
	if !IsValidKind(kind) {
		return nil, []error{fmt.Errorf("unsupported track kind: %s", kind)}
	}

	return s.loadPaths(paths, kind)
}

func (s *Service) loadPaths(paths []string, kind string) ([]*Track, []error) {
	s.loadMutex.Lock()
	defer s.loadMutex.Unlock()

	tracks := make([]*Track, 0, len(paths))
	errs := make([]error, 0)
	for _, trackPath := range paths {
		filename := filepath.Base(trackPath)
		preparedTrackPath, err := s.PrepareTrack(trackPath, s.audioBitRate())
		if err != nil {
			s.log.Warn("Failed to prepare a track for streaming: "+err.Error(), "track", filename)
			errs = append(errs, fmt.Errorf("%s: %w", filename, err))
			continue
		}

		track, err := s.AddTrack(filename, preparedTrackPath, kind, 0, 0, 0)
		if err != nil {
			s.log.Warn("Failed to save track to database: "+err.Error(), "track", filename)
			errs = append(errs, fmt.Errorf("%s: %w", filename, err))
			_ = fs.DeleteFile(preparedTrackPath)
			continue
		}

		err = fs.DeleteFile(trackPath)
		if err != nil {
			s.log.Warn("Failed to delete original copy of prepared track: "+err.Error(), "track", filename)
		}

		tracks = append(tracks, track)
	}

	if len(tracks) > 0 {
		s.log.Info(fmt.Sprintf("Loaded %d new track(s) from disk.", len(tracks)))
		select {
		case s.LoadedTracksNotify <- len(tracks):
		default:
		}
	}

	return tracks, errs
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
	name = strings.ReplaceAll(name, ".flac", "")
	name = strings.ReplaceAll(name, ".m4a", "")
	name = strings.ReplaceAll(name, ".M4A", "")
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
