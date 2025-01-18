package playback

import (
	"errors"
	"math"
	"sync"
	"time"

	"github.com/cheatsnake/airstation/internal/ffmpeg"
	"github.com/cheatsnake/airstation/internal/hls"
	"github.com/cheatsnake/airstation/internal/track"
)

type State struct {
	CurrentTrack         *track.Track
	CurrentTrackPlayback float64 // Seconds
	NextTrack            *track.Track
	IsPlaying            bool
	TrackQueue           *track.Queue

	playlist       *hls.Playlist
	totalRefreshes int64
	temporaryDir   string
	refreshRate    float64 // Seconds
	mutex          sync.Mutex
}

func NewState(tq track.Queue, tmpDir string) *State {
	return &State{
		CurrentTrack:         tq.CurrentTrack(),
		CurrentTrackPlayback: 0,
		NextTrack:            tq.NextTrack(),
		IsPlaying:            false,
		TrackQueue:           &tq,

		totalRefreshes: 0,
		temporaryDir:   tmpDir,
		refreshRate:    1,
	}
}

func (s *State) Run() {
	ticker := time.NewTicker(time.Duration(s.refreshRate) * time.Second)
	defer ticker.Stop()

	for range ticker.C {

		if s.IsPlaying {
			s.mutex.Lock()

			s.CurrentTrackPlayback += s.refreshRate
			s.totalRefreshes++

			// every time a new segment is played
			if math.Mod(float64(s.totalRefreshes), float64(s.playlist.MaxSegmentDuration)) == 0 {
				s.playlist.UpdateMediaSequence()
			}

			s.playlist.UpdateDisconSequence(s.CurrentTrackPlayback)

			if s.CurrentTrackPlayback >= s.CurrentTrack.Duration {
				s.loadNextTrack()
			}

			s.mutex.Unlock()
		}
	}
}

func (s *State) TogglePlaying() error {
	if s.CurrentTrack == nil {
		return errors.New("no tracks for playing")
	}

	if s.playlist == nil {
		s.initHLSPlaylist()
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.IsPlaying = !s.IsPlaying

	return nil
}

func (s *State) GenerateHLSPlaylist() string {
	pl := s.playlist.Generate(s.CurrentTrackPlayback)
	return pl
}

func (s *State) initHLSPlaylist() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cur := s.currentTrackSegments()
	next := s.nextTrackSegments()
	s.playlist = hls.NewPlaylist(cur, next)
}

func (s *State) currentTrackSegments() []*hls.Segment {
	if s.CurrentTrack == nil {
		return nil
	}

	err := ffmpeg.GenerateHLSPlaylist(s.CurrentTrack.Path, s.temporaryDir, s.CurrentTrack.ID, hls.DefaultMaxSegmentDuration)
	if err != nil {
		panic(err)
	}

	currentTrackSegments := hls.GenerateSegments(
		s.CurrentTrack.Duration,
		hls.DefaultMaxSegmentDuration,
		s.CurrentTrack.ID,
		s.temporaryDir,
	)

	return currentTrackSegments
}

func (s *State) nextTrackSegments() []*hls.Segment {
	if s.NextTrack == nil {
		return nil
	}

	err := ffmpeg.GenerateHLSPlaylist(s.NextTrack.Path, s.temporaryDir, s.NextTrack.ID, hls.DefaultMaxSegmentDuration)
	if err != nil {
		panic(err)
	}

	nextTrackSegments := hls.GenerateSegments(
		s.NextTrack.Duration,
		hls.DefaultMaxSegmentDuration,
		s.NextTrack.ID,
		s.temporaryDir,
	)

	return nextTrackSegments
}

func (s *State) loadNextTrack() {
	s.CurrentTrackPlayback = 0
	s.TrackQueue.Spin()
	s.CurrentTrack = s.TrackQueue.CurrentTrack()
	s.NextTrack = s.TrackQueue.NextTrack()
	s.playlist.Next(s.nextTrackSegments())
}
