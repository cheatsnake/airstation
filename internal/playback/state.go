// Package playback manages audio playback state, track transitions, and HLS playlist generation.
// It coordinates the timing and sequencing of audio tracks, maintaining synchronized state
// for streaming playback, including current position, play/pause control, and playlist updates.
// This package interacts with the track service to load tracks, generate segments, and handle
// queue changes in a thread-safe manner.
package playback

import (
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cheatsnake/airstation/internal/live"
	"github.com/cheatsnake/airstation/internal/pkg/hls"
	"github.com/cheatsnake/airstation/internal/queue"
	"github.com/cheatsnake/airstation/internal/track"
	"github.com/cheatsnake/airstation/internal/voiceover"
)

// State represents the current playback state of the application, including the currently playing track,
// elapsed playback time, playlist management, and synchronization tools for safe concurrent access.
type State struct {
	CurrentTrack        *track.Track         `json:"currentTrack"`               // The currently playing track
	CurrentTrackElapsed float64              `json:"currentTrackElapsed"`        // Seconds elapsed since the current track started playing
	CurrentVoiceOver    *voiceover.VoiceOver `json:"currentVoiceOver,omitempty"` // Voice-over mixed into the current track, if any
	NextTrack           *track.Track         `json:"nextTrack,omitempty"`        // The next track in the schedule
	LiveMode            bool                 `json:"liveMode"`                   // Whether the station is currently on a live source
	LiveStartedAt       int64                `json:"liveStartedAt,omitempty"`    // Unix time the live source started
	IsPlaying           bool                 `json:"isPlaying"`                  // Whether a track is currently playing
	Mode                string               `json:"mode"`                       // Playout mode (automation for now)
	UpdatedAt           int64                `json:"updatedAt"`                  // Unix timestamp of the last state update

	NewTrackNotify chan string `json:"-"` // Channel to notify when a new track starts playing
	PlayNotify     chan bool   `json:"-"` // Channel to notify when playback starts
	PauseNotify    chan bool   `json:"-"` // Channel to notify when playback is paused

	PlaylistStr string        `json:"-"` // Current HLS playlist as a string
	playlist    *hls.Playlist // Internal representation of the HLS playlist
	playlistDir string        // Directory where HLS playlist segments are stored

	refreshCount    int64   // Number of state refresh cycles completed
	refreshInterval float64 // Time interval (in seconds) between state updates

	trackService     *track.Service
	queueService     *queue.Service
	playbackService  *Service
	voiceoverService *voiceover.Service
	liveManager      *live.Manager

	log   *slog.Logger
	mutex sync.Mutex
}

// NewState creates and initializes a new playback State instance.
func NewState(ts *track.Service, qs *queue.Service, ps *Service, vos *voiceover.Service, lm *live.Manager, tmpDir string, log *slog.Logger) *State {
	return &State{
		CurrentTrack:        nil,
		CurrentTrackElapsed: 0,
		CurrentVoiceOver:    nil,
		NextTrack:           nil,
		IsPlaying:           false,
		Mode:                "automation",
		UpdatedAt:           time.Now().Unix(),

		NewTrackNotify: make(chan string),
		PlayNotify:     make(chan bool),
		PauseNotify:    make(chan bool),

		trackService:     ts,
		queueService:     qs,
		playbackService:  ps,
		voiceoverService: vos,
		liveManager:      lm,

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

		if s.LiveMode && s.liveManager != nil {
			// Keep the previous playlist on air until FFmpeg has produced its
			// first live segment, otherwise players stall during takeover.
			if playlist := s.liveManager.BuildPlaylist(); playlist != "" {
				s.PlaylistStr = playlist
			}
			s.CurrentTrackElapsed = time.Since(time.Unix(s.LiveStartedAt, 0)).Seconds()
			s.UpdatedAt = time.Now().Unix()
			s.mutex.Unlock()
			continue
		}

		if s.CurrentTrack == nil || s.playlist == nil {
			s.mutex.Unlock()
			continue
		}

		s.CurrentTrackElapsed += s.refreshInterval
		s.refreshCount++

		if s.CurrentTrackElapsed >= s.CurrentTrack.Duration {
			err := s.loadNextTrack()
			if err != nil {
				s.log.Error(err.Error())
			}

			if s.CurrentTrack != nil {
				go s.queueService.CleanupHLSPlaylists(s.playlistDir)
				go s.playbackService.AddPlaybackHistory(s.CurrentTrack.Name, s.CurrentTrack.Kind)
			}
		}

		if s.playlist != nil && s.CurrentTrack != nil {
			s.PlaylistStr = s.playlist.Generate(s.CurrentTrackElapsed)
		}
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

	currentVoiceOver, err := s.initHLSPlaylist(current, next)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	s.CurrentTrack = current
	s.CurrentVoiceOver = currentVoiceOver
	s.NextTrack = next
	s.PlaylistStr = s.playlist.Generate(s.CurrentTrackElapsed)
	s.UpdatedAt = time.Now().Unix()
	s.IsPlaying = true
	s.Mode = "automation"
	s.mutex.Unlock()

	s.PlayNotify <- true
	go s.playbackService.AddPlaybackHistory(current.Name, current.Kind)

	return nil
}

// Pause stops playback, clears current playback state and playlist.
func (s *State) Pause() {
	s.mutex.Lock()
	s.CurrentTrack = nil
	s.CurrentTrackElapsed = 0
	s.CurrentVoiceOver = nil
	s.NextTrack = nil
	s.playlist = nil
	s.PlaylistStr = ""
	s.IsPlaying = false
	s.UpdatedAt = time.Now().Unix()
	s.mutex.Unlock()

	s.PauseNotify <- false
}

// Restart restarts the current queue item from the beginning. It is used when
// an operator wants to apply a voice-over to the track that is already on air.
func (s *State) Restart() error {
	s.Pause()
	return s.Play()
}

// StartLiveIngest switches the station from automation to a microphone-driven
// live session. The caller feeds the WebM/Opus body through WriteLiveIngest and
// calls StopLive when the body ends. The queue is left untouched so automation
// can resume afterwards.
func (s *State) StartLiveIngest() error {
	if s.liveManager == nil {
		return errors.New("live manager is not configured")
	}
	if err := s.liveManager.StartIngest(); err != nil {
		return err
	}

	return s.startLive("Live (on air)")
}

// StartLiveExternal switches to a live source fed by an external stream URL.
func (s *State) StartLiveExternal(url string) error {
	if s.liveManager == nil {
		return errors.New("live manager is not configured")
	}
	if err := s.liveManager.StartExternal(url); err != nil {
		return err
	}

	return s.startLive("Live (external)")
}

// startLive switches the playback state to the live source without emitting a
// pause event. Listeners keep hearing the previous automation playlist until
// the first live segment is available, which avoids a gap while FFmpeg starts.
func (s *State) startLive(name string) error {
	s.mutex.Lock()
	previousPlaylist := s.PlaylistStr
	s.LiveMode = true
	s.LiveStartedAt = time.Now().Unix()
	startedAt := s.LiveStartedAt
	s.CurrentTrack = &track.Track{
		ID:       "live",
		Name:     name,
		Path:     "",
		Duration: 86400,
		Kind:     "live",
	}
	s.CurrentVoiceOver = nil
	s.NextTrack = nil
	s.CurrentTrackElapsed = 0
	s.playlist = nil
	// Keep the last automation playlist on air until the first live segment is
	// ready, so players do not stall during the takeover.
	s.PlaylistStr = previousPlaylist
	s.IsPlaying = true
	s.Mode = "live"
	s.UpdatedAt = time.Now().Unix()
	s.mutex.Unlock()

	go s.notifyWhenLiveReady(startedAt)
	return nil
}

// notifyWhenLiveReady waits until FFmpeg has produced a playable live playlist
// and only then asks players to reload /stream. Reloading earlier would attach
// them to the old automation playlist and a later media-sequence reset could
// leave the player stalled.
func (s *State) notifyWhenLiveReady(startedAt int64) {
	if s.liveManager == nil {
		return
	}

	deadline := time.Now().Add(15 * time.Second)
	for {
		s.mutex.Lock()
		stillLive := s.LiveMode && s.LiveStartedAt == startedAt
		s.mutex.Unlock()

		if !stillLive {
			return
		}

		if playlist := s.liveManager.BuildPlaylist(); playlist != "" {
			s.mutex.Lock()
			if s.LiveMode && s.LiveStartedAt == startedAt {
				s.PlaylistStr = playlist
				s.UpdatedAt = time.Now().Unix()
			}
			stillLive = s.LiveMode && s.LiveStartedAt == startedAt
			s.mutex.Unlock()

			if stillLive {
				s.PlayNotify <- true
			}
			return
		}

		if time.Now().After(deadline) {
			// The source did not produce audio in time; let players switch to
			// whatever /stream currently serves instead of waiting forever.
			s.mutex.Lock()
			stillLive = s.LiveMode && s.LiveStartedAt == startedAt
			s.mutex.Unlock()
			if stillLive {
				s.PlayNotify <- true
			}
			return
		}

		time.Sleep(200 * time.Millisecond)
	}
}

// WriteLiveIngest feeds the open ingest body into the live encoder. It blocks
// until the body ends or the encoder exits.
func (s *State) WriteLiveIngest(r io.Reader) (int64, error) {
	if s.liveManager == nil {
		return 0, errors.New("live manager is not configured")
	}
	return s.liveManager.Write(r)
}

// WriteLiveIngestBytes feeds one binary ingest message into the live encoder.
func (s *State) WriteLiveIngestBytes(p []byte) (int, error) {
	if s.liveManager == nil {
		return 0, errors.New("live manager is not configured")
	}
	return s.liveManager.WriteBytes(p)
}

// StopLive ends the live source and resumes automation if a queue is available.
// It is safe to call more than once; only the first call has an effect.
func (s *State) StopLive() error {
	if s.liveManager == nil {
		return errors.New("live manager is not configured")
	}

	s.mutex.Lock()
	if !s.LiveMode {
		s.mutex.Unlock()
		return nil
	}
	s.LiveMode = false
	s.mutex.Unlock()

	// Stop the encoder but keep the last live playlist and segments readable.
	// The player continues on them while automation is rendered instead of
	// receiving an empty /stream, which would leave hls.js retrying a dead
	// manifest.
	liveSession := s.liveManager.SessionID()
	s.liveManager.Stop()

	s.mutex.Lock()
	s.LiveStartedAt = 0
	s.CurrentTrackElapsed = 0
	s.mutex.Unlock()

	// Resume automation if possible. Play() replaces PlaylistStr with the fresh
	// automation playlist and then sends the play event, so listeners reload
	// only once there is something new to play. Emitting a pause event here
	// would race with that reload and can leave the player stalled until it is
	// manually restarted.
	if err := s.Play(); err != nil {
		// Queue may be empty; clear the now-dead stream and tell players.
		s.log.Debug("Automation did not resume after live: " + err.Error())
		s.mutex.Lock()
		s.CurrentTrack = nil
		s.CurrentTrackElapsed = 0
		s.CurrentVoiceOver = nil
		s.NextTrack = nil
		s.playlist = nil
		s.PlaylistStr = ""
		s.IsPlaying = false
		s.Mode = "automation"
		s.UpdatedAt = time.Now().Unix()
		s.mutex.Unlock()
		s.liveManager.CleanupSession(liveSession)
		s.PauseNotify <- false
		return nil
	}

	// Give listeners a moment to process the play event and detach from the
	// live playlist before its files are removed. CleanupSession ignores the
	// call if the operator has already started a new live session.
	go func() {
		time.Sleep(5 * time.Second)
		s.liveManager.CleanupSession(liveSession)
	}()

	return nil
}

// IsLive reports whether a live source is currently on air.
func (s *State) IsLive() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.LiveMode
}

// LiveManagerMode reports the active live source mode (mic/external/empty).
func (s *State) LiveManagerMode() string {
	if s.liveManager == nil {
		return ""
	}
	return s.liveManager.Mode()
}

// LiveSegmentCount reports how many segments are in the current live window.
func (s *State) LiveSegmentCount() int {
	if s.liveManager == nil {
		return 0
	}
	return s.liveManager.SegmentCount()
}

// LiveStartedUnix returns the Unix time the live session began, or 0.
func (s *State) LiveStartedUnix() int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.LiveStartedAt
}

// CurrentTrackID returns the ID of the track currently on air, if any.
func (s *State) CurrentTrackID() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.CurrentTrack == nil {
		return ""
	}
	return s.CurrentTrack.ID
}

// PlaylistSnapshot returns the current HLS playlist and whether playback is
// active. Callers must not read PlaylistStr directly: it is replaced under the
// playback mutex and an unsynchronized read can observe a partial manifest
// while the station switches between live and automation.
func (s *State) PlaylistSnapshot() (string, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.PlaylistStr, s.IsPlaying
}

// CurrentTrackName returns the name of the track currently on air, or an empty
// string. It is safe to call from event notifiers.
func (s *State) CurrentTrackName() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.CurrentTrack == nil {
		return ""
	}
	return s.CurrentTrack.Name
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

	isCurrentTrackChanged := current != nil && (s.CurrentTrack == nil || s.CurrentTrack.ID != current.ID)
	if isCurrentTrackChanged { // Restart if current track changed
		s.Pause()
		return s.Play()
	}

	if next == nil {
		s.mutex.Lock()
		s.NextTrack = nil
		s.mutex.Unlock()
		return nil
	}

	needsNextRender := s.playlist == nil || s.playlist.FirstNextTrackSegment() == nil
	var nextSegment *hls.Segment
	if !needsNextRender {
		nextSegment = s.playlist.FirstNextTrackSegment()
		nextPath := filepath.Base(nextSegment.Path)
		if !strings.Contains(nextPath, next.ID) {
			needsNextRender = true
		}
		if s.voiceoverService != nil {
			if s.voiceoverService.HasPendingForNext() {
				needsNextRender = true
			} else if strings.Contains(nextPath, next.ID+"-vo") && !s.voiceoverService.IsRenderedFor(next.ID) {
				// The operator cancelled a voice-over that was already mixed
				// into the upcoming track. Re-render the clean version.
				needsNextRender = true
			}
		}
	}

	if needsNextRender {
		nextSegments, _, err := s.makeHLSSegments(next, s.playlistDir, "next")
		if err != nil {
			return err
		}
		s.mutex.Lock()
		s.playlist.ChangeNext(nextSegments)
		s.NextTrack = next
		s.mutex.Unlock()
	} else {
		s.mutex.Lock()
		s.NextTrack = next
		s.mutex.Unlock()
	}

	return nil
}

// initHLSPlaylist prepares HLS segments for the current and next tracks,
// initializing a new playlist. It returns the voice-over applied to the current
// track (if any) and consumes that voice-over before the next track is
// prepared, so a one-shot link does not repeat on the following rotation.
func (s *State) initHLSPlaylist(current, next *track.Track) (*voiceover.VoiceOver, error) {
	currentSeg, currentVoiceOver, err := s.makeHLSSegments(current, s.playlistDir, "current")
	if err != nil {
		return nil, err
	}

	if currentVoiceOver != nil && s.voiceoverService != nil {
		s.voiceoverService.MarkPlayed(currentVoiceOver.ID)
	}

	nextSeg, _, err := s.makeHLSSegments(next, s.playlistDir, "next")
	if err != nil {
		return nil, err
	}

	s.mutex.Lock()
	s.playlist = hls.NewPlaylist(currentSeg, nextSeg)
	s.UpdatedAt = time.Now().Unix()
	s.mutex.Unlock()

	return currentVoiceOver, nil
}

// loadNextTrack advances the queue, resets elapsed time, and updates playlist with next segments.
func (s *State) loadNextTrack() error {
	s.CurrentTrackElapsed = 0

	if err := s.queueService.SpinQueue(); err != nil {
		return err
	}

	current, next, err := s.queueService.CurrentAndNextTrack()
	if err != nil {
		return err
	}

	if current == nil {
		s.CurrentTrack = nil
		s.CurrentVoiceOver = nil
		s.NextTrack = nil
		s.playlist = nil
		s.PlaylistStr = ""
		s.IsPlaying = false
		return nil
	}

	s.CurrentTrack = current
	s.CurrentVoiceOver = s.renderedVoiceOver(current)
	s.NextTrack = next

	if s.CurrentVoiceOver != nil && s.voiceoverService != nil {
		// The mix already aired as part of the current segment set; consume
		// it so the next rotation of this track is clean.
		s.voiceoverService.MarkPlayed(s.CurrentVoiceOver.ID)
	}

	nextTrackSegments, _, err := s.makeHLSSegments(next, s.playlistDir, "next")
	if err != nil {
		return err
	}

	s.NewTrackNotify <- current.Name
	s.playlist.Next(nextTrackSegments)
	return nil
}

// makeHLSSegments generates HLS segments for a given track. If the track has
// an RJ voice-over plan, the voice is mixed into the music bed first and the
// mixed file is segmented instead.
func (s *State) makeHLSSegments(t *track.Track, dir string, role string) ([]*hls.Segment, *voiceover.VoiceOver, error) {
	if t == nil {
		return []*hls.Segment{}, nil, nil
	}

	useTrack := t
	var appliedVoiceOver *voiceover.VoiceOver
	if s.voiceoverService != nil {
		mixed, vo, err := s.voiceoverService.Prepare(t, role, dir)
		if err != nil {
			s.log.Warn("Voice-over mix failed, playing the original track: " + err.Error())
		}
		if mixed != nil {
			useTrack = mixed
			appliedVoiceOver = vo
		}
	}

	err := s.trackService.MakeHLSPlaylist(useTrack.Path, dir, useTrack.ID, hls.DefaultMaxSegmentDuration)
	if err != nil {
		return nil, nil, err
	}

	segments := hls.GenerateSegments(
		useTrack.Duration,
		hls.DefaultMaxSegmentDuration,
		useTrack.ID,
		dir,
	)

	return segments, appliedVoiceOver, nil
}

func (s *State) renderedVoiceOver(t *track.Track) *voiceover.VoiceOver {
	if s.voiceoverService == nil || t == nil {
		return nil
	}

	vo, err := s.voiceoverService.RenderedFor(t.ID)
	if err != nil {
		return nil
	}

	return vo
}
