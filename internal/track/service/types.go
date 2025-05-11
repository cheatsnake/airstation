package trackservice

import "github.com/cheatsnake/airstation/internal/track"

// TracksPage represents a paginated response containing a list of audio tracks.
type TracksPage struct {
	Tracks []*track.Track `json:"tracks"` // A slice of Track pointers returned for the current page.
	Page   int            `json:"page"`   // The current page number in the pagination result.
	Limit  int            `json:"limit"`  // The maximum number of tracks per page.
	Total  int            `json:"total"`  // The total number of tracks matching the query.
}

type TrackIDs struct {
	IDs []string `json:"ids"`
}
