package trackservice

import "github.com/cheatsnake/airstation/internal/track"

type TracksPage struct {
	Tracks []*track.Track `json:"tracks"`
	Page   int            `json:"page"`
	Limit  int            `json:"limit"`
	Total  int            `json:"total"`
}
