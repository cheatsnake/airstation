package playback

import (
	"errors"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/cheatsnake/airstation/internal/ffmpeg"
	"github.com/cheatsnake/airstation/internal/hls"
	"github.com/cheatsnake/airstation/internal/track"
)

type State struct {
	CurrentTrack        *track.Track // The currently playing track
	CurrentTrackElapsed float64      // Seconds elapsed since the track started playing
	NextTrack           *track.Track // The next track in the queue
	IsPlaying           bool         // Whether the track is currently playing
	TrackQueue          *track.Queue // The queue of upcoming tracks

	playlist    *hls.Playlist // HLS playlist for streaming
	playlistDir string        // Directory for temporary playlist data

	refreshCount    int64   // Total number of state refreshes
	refreshInterval float64 // Interval in seconds between state refreshes

	ffmpegCLI *ffmpeg.CLI

	log   *slog.Logger
	mutex sync.Mutex
}

func NewState(tq track.Queue, tmpDir string, ffmpegCLI *ffmpeg.CLI, log *slog.Logger) *State {
	return &State{
		CurrentTrack:        tq.CurrentTrack(),
		CurrentTrackElapsed: 0,
		NextTrack:           tq.NextTrack(),
		IsPlaying:           false,
		TrackQueue:          &tq,

		refreshCount:    0,
		playlistDir:     tmpDir,
		refreshInterval: 1,

		ffmpegCLI: ffmpegCLI,
		log:       log,
	}
}

func (s *State) Run() {
	ticker := time.NewTicker(time.Duration(s.refreshInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {

		if s.IsPlaying {
			s.mutex.Lock()

			s.CurrentTrackElapsed += s.refreshInterval
			s.refreshCount++

			// every time a new segment is played
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
}

func (s *State) TogglePlaying() error {
	if s.CurrentTrack == nil {
		return errors.New("no tracks for playing")
	}

	if s.playlist == nil {
		err := s.initHLSPlaylist()
		if err != nil {
			return err
		}
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.IsPlaying = !s.IsPlaying

	return nil
}

func (s *State) GenerateHLSPlaylist() string {
	pl := s.playlist.Generate(s.CurrentTrackElapsed)
	return pl
}

func (s *State) initHLSPlaylist() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	current, err := s.makeHLSSegments(s.CurrentTrack, s.playlistDir)
	if err != nil {
		return err
	}

	next, err := s.makeHLSSegments(s.NextTrack, s.playlistDir)
	if err != nil {
		return err
	}

	s.playlist = hls.NewPlaylist(current, next)

	return nil
}

func (s *State) loadNextTrack() error {
	s.CurrentTrackElapsed = 0
	s.TrackQueue.Spin()
	s.CurrentTrack = s.TrackQueue.CurrentTrack()
	s.NextTrack = s.TrackQueue.NextTrack()

	nextTrackSegments, err := s.makeHLSSegments(s.NextTrack, s.playlistDir)
	if err != nil {
		return err
	}

	s.playlist.Next(nextTrackSegments)
	return nil
}

func (s *State) makeHLSSegments(track *track.Track, dir string) ([]*hls.Segment, error) {
	if track == nil {
		return []*hls.Segment{}, nil
	}

	err := s.ffmpegCLI.MakeHLSPlaylist(track.Path, dir, track.ID, hls.DefaultMaxSegmentDuration)
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
