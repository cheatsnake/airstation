package playlist

import (
	"errors"
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

func validateName(name string) error {
	if len(name) < minNameLen {
		return fmt.Errorf("name must be at least %d characters", minNameLen)
	}
	if len(name) > maxNameLen {
		return fmt.Errorf("name must be at most %d characters", maxNameLen)
	}
	return nil
}

func validateDescr(descr string) error {
	if len(descr) > maxDescrLen {
		return fmt.Errorf("description must be at most %d characters", maxDescrLen)
	}
	return nil
}

func validateTracks(trackIDs []string) error {
	if len(trackIDs) > maxTracks {
		return fmt.Errorf("playlist cannot have more than %d tracks", maxTracks)
	}

	seen := make(map[string]struct{}, len(trackIDs))
	for _, id := range trackIDs {
		if id == "" {
			return errors.New("track ID cannot be empty")
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("duplicate track ID found: %s", id)
		}
		seen[id] = struct{}{}
	}
	return nil
}
