package playlist

import "github.com/cheatsnake/airstation/internal/track"

type Playlist struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Tracks      []*track.Track `json:"tracks"`
	TrackCount  int            `json:"trackCount"`
}

type Store interface {
	AddPlaylist(name, description string, trackIDs []string) (*Playlist, error)
	Playlists() ([]*Playlist, error)
	Playlist(id string) (*Playlist, error)
	IsPlaylistExists(name string) (bool, error)
	EditPlaylist(id, name, description string, trackIDs []string) error
	DeletePlaylist(id string) error
}
