package playback

import (
	"errors"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/cheatsnake/airstation/internal/hls"
	"github.com/cheatsnake/airstation/internal/track"
	trackservice "github.com/cheatsnake/airstation/internal/track/service"
)

type State struct {
	CurrentTrack        *track.Track `json:"currentTrack"`        // The currently playing track
	CurrentTrackElapsed float64      `json:"currentTrackElapsed"` // Seconds elapsed since the track started playing
	IsPlaying           bool         `json:"isPlaying"`           // Whether the track is currently playing

	NewTrackNotify chan string `json:"-"`
	PlayNotify     chan bool   `json:"-"`
	PauseNotify    chan bool   `json:"-"`

	playlist    *hls.Playlist // HLS playlist for streaming
	playlistDir string        // Directory for temporary playlist data

	refreshCount    int64   // Total number of state refreshes
	refreshInterval float64 // Interval in seconds between state refreshes

	trackService *trackservice.Service

	log   *slog.Logger
	mutex sync.Mutex
}

func NewState(ts *trackservice.Service, tmpDir string, log *slog.Logger) *State {
	return &State{
		CurrentTrack:        nil,
		CurrentTrackElapsed: 0,
		IsPlaying:           false,

		NewTrackNotify: make(chan string),
		PlayNotify:     make(chan bool),
		PauseNotify:    make(chan bool),

		trackService: ts,

		refreshCount:    0,
		playlistDir:     tmpDir,
		refreshInterval: 1,

		log: log,
	}
}

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

		// every time a new segment is playing
		if math.Mod(float64(s.refreshCount), float64(s.playlist.MaxSegmentDuration)) == 0 {
			s.playlist.UpdateMediaSequence()
		}

		s.playlist.UpdateDisconSequence(s.CurrentTrackElapsed)

		if s.CurrentTrackElapsed >= s.CurrentTrack.Duration {
			err := s.loadNextTrack()
			if err != nil {
				s.log.Error(err.Error())
			}
		}

		s.mutex.Unlock()
	}
}

func (s *State) Play() error {
	err := s.Load()
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.IsPlaying = true
	s.PlayNotify <- true

	return nil
}

func (s *State) Pause() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.IsPlaying = false
	s.PauseNotify <- false
}

func (s *State) Load() error {
	current, next, err := s.trackService.CurrentAndNextTrack()
	if err != nil {
		return err
	}

	if current == nil {
		return errors.New("no tracks for playing")
	}

	if s.playlist == nil {
		err = s.initHLSPlaylist(current, next)
		if err != nil {
			return err
		}
	} else {
		nextSeg, err := s.makeHLSSegments(next, s.playlistDir)
		if err != nil {
			return err
		}
		s.mutex.Lock()
		s.playlist.ChangeNext(nextSeg)
		s.mutex.Unlock()
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.CurrentTrack = current

	return nil
}

func (s *State) GenerateHLSPlaylist() string {
	pl := s.playlist.Generate(s.CurrentTrackElapsed)
	return pl
}

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
	defer s.mutex.Unlock()
	s.playlist = hls.NewPlaylist(currentSeg, nextSeg)

	return nil
}

func (s *State) loadNextTrack() error {
	s.CurrentTrackElapsed = 0
	err := s.trackService.SpinQueue()
	if err != nil {
		return err
	}

	current, next, err := s.trackService.CurrentAndNextTrack()
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
