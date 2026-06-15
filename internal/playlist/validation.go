package playlist

import (
	"errors"
	"fmt"
)

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
