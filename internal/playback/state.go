// Package playback manages audio playback state, track transitions, and HLS playlist generation.
// It coordinates the timing and sequencing of audio tracks, maintaining synchronized state
// for streaming playback, including current position, play/pause control, and playlist updates.
// This package interacts with the track service to load tracks, generate segments, and handle
// queue changes in a thread-safe manner.
package playback

import (
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/cheatsnake/airstation/internal/hls"
	"github.com/cheatsnake/airstation/internal/queue"
	"github.com/cheatsnake/airstation/internal/track"
)

// State represents the current playback state of the application, including the currently playing track,
// elapsed playback time, playlist management, and synchronization tools for safe concurrent access.
type State struct {
	CurrentTrack        *track.Track `json:"currentTrack"`        // The currently playing track
	CurrentTrackElapsed float64      `json:"currentTrackElapsed"` // Seconds elapsed since the current track started playing
	IsPlaying           bool         `json:"isPlaying"`           // Whether a track is currently playing
	UpdatedAt           int64        `json:"updatedAt"`           // Unix timestamp of the last state update

	NewTrackNotify chan string `json:"-"` // Channel to notify when a new track starts playing
	PlayNotify     chan bool   `json:"-"` // Channel to notify when playback starts
	PauseNotify    chan bool   `json:"-"` // Channel to notify when playback is paused

	PlaylistStr string        `json:"-"` // Current HLS playlist as a string
	playlist    *hls.Playlist // Internal representation of the HLS playlist
	playlistDir string        // Directory where HLS playlist segments are stored

	refreshCount    int64   // Number of state refresh cycles completed
	refreshInterval float64 // Time interval (in seconds) between state updates

	trackService    *track.Service
	queueService    *queue.Service
	playbackService *Service

	log   *slog.Logger
	mutex sync.Mutex
}

// NewState creates and initializes a new playback State instance.
func NewState(ts *track.Service, qs *queue.Service, ps *Service, tmpDir string, log *slog.Logger) *State {
	return &State{
		CurrentTrack:        nil,
		CurrentTrackElapsed: 0,
		IsPlaying:           false,
		UpdatedAt:           time.Now().Unix(),

		NewTrackNotify: make(chan string),
		PlayNotify:     make(chan bool),
		PauseNotify:    make(chan bool),

		trackService:    ts,
		queueService:    qs,
		playbackService: ps,

		refreshCount:    0,
		playlistDir:     tmpDir,
		refreshInterval: 1,

		log: log,
	}
}

// Run starts the state update loop which refreshes playback progress and switches tracks when needed.
func (s *State) Run() {
	ticker := time.NewTicker(time.Duration(s.refreshInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !s.IsPlaying {
			continue
		}

		s.mutex.Lock()
		s.CurrentTrackElapsed += s.refreshInterval
		s.refreshCount++

		if s.CurrentTrackElapsed >= s.CurrentTrack.Duration {
			err := s.loadNextTrack()
			if err != nil {
				s.log.Error(err.Error())
			}

			go s.queueService.CleanupHLSPlaylists(s.playlistDir)
			go s.playbackService.AddPlaybackHistory(s.CurrentTrack.Name)
		}

		s.PlaylistStr = s.playlist.Generate(s.CurrentTrackElapsed)
		s.UpdatedAt = time.Now().Unix()
		s.mutex.Unlock()
	}
}

// Play starts playback by loading the current and next tracks into the HLS playlist.
func (s *State) Play() error {
	current, next, err := s.queueService.CurrentAndNextTrack()
	if err != nil {
		return err
	}

	if current == nil {
		return errors.New("playback queue is empty")
	}

	err = s.initHLSPlaylist(current, next)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	s.CurrentTrack = current
	s.PlaylistStr = s.playlist.Generate(s.CurrentTrackElapsed)
	s.UpdatedAt = time.Now().Unix()
	s.IsPlaying = true
	s.mutex.Unlock()

	s.PlayNotify <- true
	go s.playbackService.AddPlaybackHistory(current.Name)

	return nil
}

// Pause stops playback, clears current playback state and playlist.
func (s *State) Pause() {
	s.mutex.Lock()
	s.CurrentTrack = nil
	s.CurrentTrackElapsed = 0
	s.playlist = nil
	s.PlaylistStr = ""
	s.IsPlaying = false
	s.UpdatedAt = time.Now().Unix()
	s.mutex.Unlock()

	s.PauseNotify <- false
}

// Reload refreshes the current playlist based on updated queue state, used after queue changes.
func (s *State) Reload() error {
	if !s.IsPlaying {
		return nil
	}

	current, next, err := s.queueService.CurrentAndNextTrack()
	if err != nil {
		return err
	}

	isCurrentTrackChanged := current != nil && s.CurrentTrack.ID != current.ID
	if isCurrentTrackChanged { // Restart if current track changed
		s.Pause()
		err = s.Play()
		if err != nil {
			return err
		}
	}

	segment := s.playlist.FirstNextTrackSegment()
	isNextTrackChanged := segment != nil && !strings.Contains(segment.Path, next.ID)
	if isNextTrackChanged { // Change segments for next track if it changed
		nextSeg, err := s.makeHLSSegments(next, s.playlistDir)
		if err != nil {
			return err
		}
		s.mutex.Lock()
		s.playlist.ChangeNext(nextSeg)
		s.mutex.Unlock()
	}

	return nil
}

// initHLSPlaylist prepares HLS segments for the current and next tracks, initializing a new playlist.
func (s *State) initHLSPlaylist(current, next *track.Track) error {
	currentSeg, err := s.makeHLSSegments(current, s.playlistDir)
	if err != nil {
		return err
	}

	nextSeg, err := s.makeHLSSegments(next, s.playlistDir)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	s.playlist = hls.NewPlaylist(currentSeg, nextSeg)
	s.UpdatedAt = time.Now().Unix()
	s.mutex.Unlock()

	return nil
}

// loadNextTrack advances the queue, resets elapsed time, and updates playlist with next segments.
func (s *State) loadNextTrack() error {
	s.CurrentTrackElapsed = 0
	err := s.queueService.SpinQueue()
	if err != nil {
		return err
	}

	current, next, err := s.queueService.CurrentAndNextTrack()
	if err != nil {
		return err
	}

	s.CurrentTrack = current
	nextTrackSegments, err := s.makeHLSSegments(next, s.playlistDir)
	if err != nil {
		return err
	}

	s.NewTrackNotify <- current.Name
	s.playlist.Next(nextTrackSegments)
	return nil
}

// makeHLSSegments generates HLS segments for a given track.
func (s *State) makeHLSSegments(track *track.Track, dir string) ([]*hls.Segment, error) {
	if track == nil {
		return []*hls.Segment{}, nil
	}

	err := s.trackService.MakeHLSPlaylist(track.Path, dir, track.ID, hls.DefaultMaxSegmentDuration)
	if err != nil {
		return nil, err
	}

	segments := hls.GenerateSegments(
		track.Duration,
		hls.DefaultMaxSegmentDuration,
		track.ID,
		dir,
	)

	return segments, nil
}
