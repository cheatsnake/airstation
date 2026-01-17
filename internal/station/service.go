package station

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) Info() (*Info, error) {
	rawProps, err := s.store.StationProperties()
	if err != nil {
		return nil, err
	}

	info := &Info{}

	for _, prop := range rawProps {
		switch prop.Key {
		case propName:
			info.Name = prop.Value
		case propDescription:
			info.Description = prop.Value
		case propFaviconURL:
			info.FaviconURL = prop.Value
		case propLogoURL:
			info.LogoURL = prop.Value
		case propLocation:
			info.Location = prop.Value
		case propTimezone:
			info.Timezone = prop.Value
		case propLinks:
			info.Links = prop.Value
		case propTheme:
			info.Theme = prop.Value
		}
	}

	return info, nil
}

func (s *Service) EditInfo(editedInfo *Info) (*Info, error) {
	currentInfo, err := s.Info()
	if err != nil {
		return nil, err
	}

	if currentInfo.Name != editedInfo.Name {
		if _, err := s.store.UpsertStationProperty(propName, editedInfo.Name); err != nil {
			return nil, err
		}
	}

	if currentInfo.Description != editedInfo.Description {
		if _, err := s.store.UpsertStationProperty(propDescription, editedInfo.Description); err != nil {
			return nil, err
		}
	}

	if currentInfo.FaviconURL != editedInfo.FaviconURL {
		if _, err := s.store.UpsertStationProperty(propFaviconURL, editedInfo.FaviconURL); err != nil {
			return nil, err
		}
	}

	if currentInfo.LogoURL != editedInfo.LogoURL {
		if _, err := s.store.UpsertStationProperty(propLogoURL, editedInfo.LogoURL); err != nil {
			return nil, err
		}
	}

	if currentInfo.Location != editedInfo.Location {
		if _, err := s.store.UpsertStationProperty(propLocation, editedInfo.Location); err != nil {
			return nil, err
		}
	}

	if currentInfo.Timezone != editedInfo.Timezone {
		if _, err := s.store.UpsertStationProperty(propTimezone, editedInfo.Timezone); err != nil {
			return nil, err
		}
	}

	if currentInfo.Links != editedInfo.Links {
		if _, err := s.store.UpsertStationProperty(propLinks, editedInfo.Links); err != nil {
			return nil, err
		}
	}

	if currentInfo.Theme != editedInfo.Theme {
		if _, err := s.store.UpsertStationProperty(propTheme, editedInfo.Theme); err != nil {
			return nil, err
		}
	}

	freshInfo, err := s.Info()
	if err != nil {
		return nil, err
	}

	return freshInfo, nil
}
