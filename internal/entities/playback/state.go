package playback

import (
	"sync"
	"time"

	"github.com/cheatsnake/airstation/internal/entities/hls"
)

type State struct {
	CurrentTrackID string
	ElapsedSeconds int
	NextTrackID    string
	IsPlaying      bool

	Playlist hls.Playlist

	LastUpdateTime time.Time

	Mutex sync.Mutex
}
