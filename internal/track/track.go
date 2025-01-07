package track

import "github.com/cheatsnake/airstation/internal/tools/ulid"

type Track struct {
	ID       string
	Name     string
	Path     string
	Duration float64 // Seconds
	Bitrate  int     // Kbps
}

func New(name, path string, duration float64, bitrate int) *Track {
	return &Track{
		ID:       ulid.New(),
		Name:     name,
		Path:     path,
		Duration: duration,
		Bitrate:  bitrate,
	}
}
