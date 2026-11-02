package playback

type History struct {
	ID        int    `json:"id"`
	PlayedAt  int64  `json:"playedAt"`
	TrackName string `json:"trackName"`
	Kind      string `json:"kind"`
}

type Store interface {
	AddPlaybackHistory(playedAt int64, trackName, kind string) error
	RecentPlaybackHistory(limit int) ([]*History, error)
	DeleteOldPlaybackHistory() (int64, error)
	MusicPlayedSince(ts int64) (int, error)
}
