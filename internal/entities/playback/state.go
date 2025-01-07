package playback

import (
	"errors"
	"math"
	"sync"
	"time"

	"github.com/cheatsnake/airstation/internal/entities/hls"
	"github.com/cheatsnake/airstation/internal/entities/track"
	"github.com/cheatsnake/airstation/internal/ffmpeg"
)

type State struct {
	CurrentTrack         *track.Track
	CurrentTrackPlayback float64 // Seconds
	NextTrack            *track.Track
	IsPlaying            bool
	TrackQueue           *track.Queue
	Playlist             *hls.Playlist

	TotalRefreshes     int64
	MediaSequence      int64
	DisconSequence     int64
	maxSegmentDuration int // Seconds
	maxSegmentsAmount  int
	temporaryDir       string
	refreshRate        float64 // Seconds
	mutex              sync.Mutex
}

func NewState(tq track.Queue, tmpDir string) *State {
	return &State{
		CurrentTrack:         tq.CurrentTrack(),
		CurrentTrackPlayback: 0,
		NextTrack:            tq.NextTrack(),
		IsPlaying:            false,
		TrackQueue:           &tq,

		TotalRefreshes:     0,
		MediaSequence:      0,
		DisconSequence:     0,
		maxSegmentDuration: 5.0,
		maxSegmentsAmount:  3,
		temporaryDir:       tmpDir,
		refreshRate:        1,
	}
}

func (s *State) Run() {
	ticker := time.NewTicker(time.Duration(s.refreshRate) * time.Second)
	defer ticker.Stop()

	for range ticker.C {

		if s.IsPlaying {
			s.mutex.Lock()
			s.CurrentTrackPlayback += s.refreshRate
			s.TotalRefreshes++

			if math.Mod(float64(s.TotalRefreshes), float64(s.maxSegmentDuration)) == 0 {
				s.MediaSequence++
			}

			if s.CurrentTrackPlayback >= s.CurrentTrack.Duration {
				s.CurrentTrackPlayback = 0
				s.DisconSequence++
				s.TrackQueue.Spin()
				s.CurrentTrack = s.TrackQueue.CurrentTrack()
				s.NextTrack = s.TrackQueue.NextTrack()
				s.Playlist.Next(s.nextTrackSegments())
			}

			s.mutex.Unlock()
		}
	}
}

func (s *State) TogglePlaying() error {
	if s.CurrentTrack == nil {
		return errors.New("no tracks for playing")
	}

	if s.Playlist == nil {
		s.initHLSPlaylist()
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.IsPlaying = !s.IsPlaying

	return nil
}

func (s *State) initHLSPlaylist() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cur := s.currentTrackSegments()
	next := s.nextTrackSegments()
	s.Playlist = hls.NewPlaylist(cur, next, s.maxSegmentDuration, s.maxSegmentsAmount)
}

func (s *State) currentTrackSegments() []*hls.Segment {
	if s.CurrentTrack == nil {
		return nil
	}

	err := ffmpeg.GenerateHLSPlaylist(s.CurrentTrack.Path, s.temporaryDir, s.CurrentTrack.ID, s.maxSegmentDuration)
	if err != nil {
		panic(err)
	}

	currentTrackSegments := hls.GenerateSegments(s.CurrentTrack.Duration, s.maxSegmentDuration, s.CurrentTrack.ID, s.temporaryDir)
	return currentTrackSegments
}

func (s *State) nextTrackSegments() []*hls.Segment {
	if s.NextTrack == nil {
		return nil
	}

	err := ffmpeg.GenerateHLSPlaylist(s.NextTrack.Path, s.temporaryDir, s.NextTrack.ID, s.maxSegmentDuration)
	if err != nil {
		panic(err)
	}

	nextTrackSegments := hls.GenerateSegments(s.NextTrack.Duration, s.maxSegmentDuration, s.NextTrack.ID, s.temporaryDir)
	return nextTrackSegments
}
