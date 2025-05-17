package playback

type History struct {
	ID        int    `json:"id"`
	PlayedAt  int64  `json:"playedAt"`
	TrackName string `json:"trackName"`
}

type Store interface {
	AddPlaybackHistory(playedAt int64, trackName string) error
	RecentPlaybackHistory(limit int) ([]*History, error)
	DeleteOldPlaybackHistory() (int64, error)
}
