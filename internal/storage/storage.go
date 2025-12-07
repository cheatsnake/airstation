package storage

import (
	"github.com/cheatsnake/airstation/internal/playback"
	"github.com/cheatsnake/airstation/internal/playlist"
	"github.com/cheatsnake/airstation/internal/queue"
	"github.com/cheatsnake/airstation/internal/station"
	"github.com/cheatsnake/airstation/internal/track"
)

type Storage interface {
	track.Store
	queue.Store
	playback.Store
	playlist.Store
	station.Store

	Close() error
}
