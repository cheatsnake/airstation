package track

type PlaybackHistory struct {
	ID        int    `json:"id"`
	PlayedAt  int64  `json:"playedAt"`
	TrackName string `json:"trackName"`
}
