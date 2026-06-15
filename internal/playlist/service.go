package playlist

import (
	"fmt"
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) AddPlaylist(name, description string, trackIDs []string) (*Playlist, error) {
	err := validateName(name)
	if err != nil {
		return nil, err
	}

	err = validateDescr(description)
	if err != nil {
		return nil, err
	}

	err = validateTracks(trackIDs)
	if err != nil {
		return nil, err
	}

	isExists, err := s.store.IsPlaylistExists(name)
	if err != nil {
		return nil, err
	}
	if isExists {
		return nil, fmt.Errorf("playlist with this name already exists")
	}

	pl, err := s.store.AddPlaylist(name, description, trackIDs)
	return pl, err
}

func (s *Service) Playlists() ([]*Playlist, error) {
	pls, err := s.store.Playlists()
	return pls, err
}

func (s *Service) Playlist(id string) (*Playlist, error) {
	pl, err := s.store.Playlist(id)
	return pl, err
}

func (s *Service) EditPlaylist(id, name, description string, trackIDs []string) error {
	err := validateName(name)
	if err != nil {
		return err
	}

	err = validateDescr(description)
	if err != nil {
		return err
	}

	err = validateTracks(trackIDs)
	if err != nil {
		return err
	}

	err = s.store.EditPlaylist(id, name, description, trackIDs)
	return err
}

func (s *Service) DeletePlaylist(id string) error {
	err := s.store.DeletePlaylist(id)
	return err
}
