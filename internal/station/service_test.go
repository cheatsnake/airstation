package station

import (
	"errors"
	"testing"
)

type mockStore struct {
	stationPropertiesFn      func() ([]*Property, error)
	upsertStationPropertyFn  func(key, value string) (*Property, error)
	deleteStationPropertyFn  func(key string) error
}

func (m *mockStore) StationProperties() ([]*Property, error) {
	if m.stationPropertiesFn != nil {
		return m.stationPropertiesFn()
	}
	return nil, nil
}

func (m *mockStore) UpsertStationProperty(key, value string) (*Property, error) {
	if m.upsertStationPropertyFn != nil {
		return m.upsertStationPropertyFn(key, value)
	}
	return &Property{Key: key, Value: value}, nil
}

func (m *mockStore) DeleteStationProperty(key string) error {
	if m.deleteStationPropertyFn != nil {
		return m.deleteStationPropertyFn(key)
	}
	return nil
}

func TestService_Info(t *testing.T) {
	t.Run("maps all property keys to Info fields", func(t *testing.T) {
		mock := &mockStore{
			stationPropertiesFn: func() ([]*Property, error) {
				return []*Property{
					{Key: "name", Value: "My Station"},
					{Key: "description", Value: "Great music"},
					{Key: "faviconURL", Value: "https://example.com/favicon.ico"},
					{Key: "logoURL", Value: "https://example.com/logo.png"},
					{Key: "location", Value: "Berlin"},
					{Key: "timezone", Value: "Europe/Berlin"},
					{Key: "links", Value: "https://example.com"},
					{Key: "theme", Value: "dark"},
				}, nil
			},
		}
		svc := NewService(mock)
		info, err := svc.Info()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Name != "My Station" {
			t.Errorf("expected Name %q, got %q", "My Station", info.Name)
		}
		if info.Description != "Great music" {
			t.Errorf("expected Description %q, got %q", "Great music", info.Description)
		}
		if info.FaviconURL != "https://example.com/favicon.ico" {
			t.Errorf("unexpected FaviconURL: %q", info.FaviconURL)
		}
		if info.LogoURL != "https://example.com/logo.png" {
			t.Errorf("unexpected LogoURL: %q", info.LogoURL)
		}
		if info.Location != "Berlin" {
			t.Errorf("unexpected Location: %q", info.Location)
		}
		if info.Timezone != "Europe/Berlin" {
			t.Errorf("unexpected Timezone: %q", info.Timezone)
		}
		if info.Links != "https://example.com" {
			t.Errorf("unexpected Links: %q", info.Links)
		}
		if info.Theme != "dark" {
			t.Errorf("unexpected Theme: %q", info.Theme)
		}
	})

	t.Run("empty properties returns zero-value Info", func(t *testing.T) {
		mock := &mockStore{
			stationPropertiesFn: func() ([]*Property, error) {
				return []*Property{}, nil
			},
		}
		svc := NewService(mock)
		info, err := svc.Info()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Name != "" || info.Theme != "" {
			t.Errorf("expected zero-value Info, got Name=%q Theme=%q", info.Name, info.Theme)
		}
	})

	t.Run("nil properties returns zero-value Info", func(t *testing.T) {
		mock := &mockStore{
			stationPropertiesFn: func() ([]*Property, error) {
				return nil, nil
			},
		}
		svc := NewService(mock)
		info, err := svc.Info()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Name != "" {
			t.Errorf("expected empty Name, got %q", info.Name)
		}
	})

	t.Run("propagates store error", func(t *testing.T) {
		mock := &mockStore{
			stationPropertiesFn: func() ([]*Property, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewService(mock)
		_, err := svc.Info()
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestService_EditInfo(t *testing.T) {
	t.Run("upserts only changed fields", func(t *testing.T) {
		var upsertedKeys []string
		mock := &mockStore{
			stationPropertiesFn: func() ([]*Property, error) {
				return []*Property{
					{Key: "name", Value: "Old Name"},
					{Key: "description", Value: "Old Desc"},
				}, nil
			},
			upsertStationPropertyFn: func(key, value string) (*Property, error) {
				upsertedKeys = append(upsertedKeys, key)
				return &Property{Key: key, Value: value}, nil
			},
		}
		svc := NewService(mock)
		_, err := svc.EditInfo(&Info{
			Name:        "New Name",
			Description: "Old Desc",
			Theme:       "dark",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(upsertedKeys) != 2 {
			t.Errorf("expected 2 upserts, got %d: %v", len(upsertedKeys), upsertedKeys)
		}
	})

	t.Run("upserts all fields when all differ", func(t *testing.T) {
		var upsertCount int
		mock := &mockStore{
			stationPropertiesFn: func() ([]*Property, error) {
				return []*Property{}, nil
			},
			upsertStationPropertyFn: func(key, value string) (*Property, error) {
				upsertCount++
				return &Property{Key: key, Value: value}, nil
			},
		}
		svc := NewService(mock)
		_, err := svc.EditInfo(&Info{
			Name:        "Name",
			Description: "Desc",
			FaviconURL:  "https://example.com/favicon.ico",
			LogoURL:     "https://example.com/logo.png",
			Location:    "Paris",
			Timezone:    "Europe/Paris",
			Links:       "https://example.com",
			Theme:       "light",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if upsertCount != 8 {
			t.Errorf("expected 8 upserts, got %d", upsertCount)
		}
	})

	t.Run("no upserts when nothing changed", func(t *testing.T) {
		var upsertCount int
		mock := &mockStore{
			stationPropertiesFn: func() ([]*Property, error) {
				return []*Property{
					{Key: "name", Value: "Same Name"},
					{Key: "theme", Value: "dark"},
				}, nil
			},
			upsertStationPropertyFn: func(key, value string) (*Property, error) {
				upsertCount++
				return &Property{Key: key, Value: value}, nil
			},
		}
		svc := NewService(mock)
		_, err := svc.EditInfo(&Info{
			Name:  "Same Name",
			Theme: "dark",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if upsertCount != 0 {
			t.Errorf("expected 0 upserts, got %d", upsertCount)
		}
	})

	t.Run("returns fresh info after edit", func(t *testing.T) {
		mock := &mockStore{
			stationPropertiesFn: func() ([]*Property, error) {
				return []*Property{
					{Key: "name", Value: "Updated Name"},
				}, nil
			},
		}
		svc := NewService(mock)
		info, err := svc.EditInfo(&Info{Name: "Updated Name"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Name != "Updated Name" {
			t.Errorf("expected Name %q, got %q", "Updated Name", info.Name)
		}
	})

	t.Run("propagates store error from Info", func(t *testing.T) {
		mock := &mockStore{
			stationPropertiesFn: func() ([]*Property, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewService(mock)
		_, err := svc.EditInfo(&Info{Name: "New Name"})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("propagates store error from Upsert", func(t *testing.T) {
		mock := &mockStore{
			stationPropertiesFn: func() ([]*Property, error) {
				return []*Property{}, nil
			},
			upsertStationPropertyFn: func(key, value string) (*Property, error) {
				return nil, errors.New("upsert failed")
			},
		}
		svc := NewService(mock)
		_, err := svc.EditInfo(&Info{Name: "New Name"})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
