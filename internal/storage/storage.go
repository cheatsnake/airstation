package storage

import (
	"github.com/cheatsnake/airstation/internal/playback"
	"github.com/cheatsnake/airstation/internal/playlist"
	"github.com/cheatsnake/airstation/internal/queue"
	"github.com/cheatsnake/airstation/internal/schedule"
	"github.com/cheatsnake/airstation/internal/station"
	"github.com/cheatsnake/airstation/internal/track"
	"github.com/cheatsnake/airstation/internal/voiceover"
)

type Storage interface {
	track.Store
	queue.Store
	playback.Store
	playlist.Store
	station.Store
	voiceover.Store
	schedule.Store

	Close() error
}
