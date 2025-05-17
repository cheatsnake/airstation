package track

// Track represents an audio track with its associated metadata.
type Track struct {
	ID       string  `json:"id"`       // A unique identifier for the track, typically generated using ULID.
	Name     string  `json:"name"`     // The name of the audio track.
	Path     string  `json:"path"`     // The file path of the audio track.
	Duration float64 `json:"duration"` // The duration of the audio track in seconds.
	BitRate  int     `json:"bitRate"`  // The bit rate of the audio track in kilobits per second (kbps).
}

type Store interface {
	Tracks(page, limit int, search, sortBy, sortOrder string) ([]*Track, int, error)
	TrackByID(ID string) (*Track, error)
	TracksByIDs(IDs []string) ([]*Track, error)
	AddTrack(name, path string, duration float64, bitRate int) (*Track, error)
	DeleteTracks(IDs []string) error
	EditTrack(track *Track) (*Track, error)
}

// Page represents a paginated response containing a list of audio tracks.
type Page struct {
	Tracks []*Track `json:"tracks"` // A slice of Track pointers returned for the current page.
	Page   int      `json:"page"`   // The current page number in the pagination result.
	Limit  int      `json:"limit"`  // The maximum number of tracks per page.
	Total  int      `json:"total"`  // The total number of tracks matching the query.
}

type BodyWithIDs struct {
	IDs []string `json:"ids"`
}
