// Package trackservice provides services related to audio track management.
package trackservice

import (
	"fmt"
	"log/slog"
	"math"
	"path"
	"path/filepath"
	"strings"
	"time"

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
func New(store storage.TrackStore, ffmpegCLI *ffmpeg.CLI, log *slog.Logger) *Service {
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
func (s *Service) AddTrack(name, path string) (*track.Track, error) {
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

func (s *Service) PrepareTrack(filePath string) (string, error) {
	newPath := replaceExtension(filePath, ".m4a")
	err := s.ffmpegCLI.ConvertAudioToAAC(filePath, newPath, 192)
	if err != nil {
		return "", err
	}

	return newPath, nil
}

func (s *Service) Tracks(page, limit int, search, sortBy, sortOrder string) (*TracksPage, error) {
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

func (s *Service) LoadTracksFromDisk(tracksDir string) ([]*track.Track, error) {
	tracks := make([]*track.Track, 0)

	mp3Filenames, err := fs.ListFilesFromDir(tracksDir, "mp3")
	if err != nil {
		return tracks, err
	}

	aacFilenames, err := fs.ListFilesFromDir(tracksDir, "aac")
	if err != nil {
		return tracks, err
	}

	trackFilenames := make([]string, 0, len(mp3Filenames)+len(aacFilenames))
	trackFilenames = append(trackFilenames, mp3Filenames...)
	trackFilenames = append(trackFilenames, aacFilenames...)

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

func (s *Service) AddPlaybackHistory(trackName string) {
	err := s.store.AddPlaybackHistory(time.Now().Unix(), trackName)
	if err != nil {
		s.log.Error("Failed to add playback history: " + err.Error())
	}
}

func (s *Service) RecentPlaybackHistory() ([]*track.PlaybackHistory, error) {
	history, err := s.store.RecentPlaybackHistory()
	return history, err
}

func (s *Service) DeleteOldPlaybackHistory() {
	_, err := s.store.DeleteOldPlaybackHistory()
	if err != nil {
		s.log.Warn("Failed to delete old playback history: " + err.Error())
	}
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
