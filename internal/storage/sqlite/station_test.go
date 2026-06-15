package sqlite

import (
	"testing"
)

func TestStationStore_UpsertAndGetProperties(t *testing.T) {
	inst := setupTestDB(t)

	t.Run("get empty properties for fresh db", func(t *testing.T) {
		props, err := inst.StationStore.StationProperties()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(props) != 0 {
			t.Errorf("expected 0 properties, got %d", len(props))
		}
	})

	t.Run("upsert new property", func(t *testing.T) {
		prop, err := inst.StationStore.UpsertStationProperty("name", "My Station")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prop.Key != "name" || prop.Value != "My Station" {
			t.Errorf("unexpected property: %v", prop)
		}
	})

	t.Run("upsert overwrites existing property", func(t *testing.T) {
		inst.StationStore.UpsertStationProperty("name", "Old Name")
		prop, err := inst.StationStore.UpsertStationProperty("name", "New Name")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prop.Value != "New Name" {
			t.Errorf("expected updated value %q, got %q", "New Name", prop.Value)
		}

		props, _ := inst.StationStore.StationProperties()
		if len(props) != 1 {
			t.Errorf("expected still 1 property, got %d", len(props))
		}
	})

	t.Run("upsert multiple properties", func(t *testing.T) {
		inst2 := setupTestDB(t)
		inst2.StationStore.UpsertStationProperty("name", "Station")
		inst2.StationStore.UpsertStationProperty("theme", "dark")
		inst2.StationStore.UpsertStationProperty("location", "Berlin")

		props, err := inst2.StationStore.StationProperties()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(props) != 3 {
			t.Errorf("expected 3 properties, got %d", len(props))
		}
	})

	t.Run("upsert with empty key returns error", func(t *testing.T) {
		_, err := inst.StationStore.UpsertStationProperty("", "value")
		if err == nil {
			t.Error("expected error for empty key, got nil")
		}
	})
}

func TestStationStore_DeleteProperty(t *testing.T) {
	inst := setupTestDB(t)

	t.Run("delete existing property", func(t *testing.T) {
		inst.StationStore.UpsertStationProperty("name", "My Station")
		err := inst.StationStore.DeleteStationProperty("name")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		props, _ := inst.StationStore.StationProperties()
		if len(props) != 0 {
			t.Errorf("expected 0 properties after delete, got %d", len(props))
		}
	})

	t.Run("delete nonexistent property returns error", func(t *testing.T) {
		err := inst.StationStore.DeleteStationProperty("nonexistent")
		if err == nil {
			t.Error("expected error deleting nonexistent property, got nil")
		}
	})

	t.Run("delete with empty key returns error", func(t *testing.T) {
		err := inst.StationStore.DeleteStationProperty("")
		if err == nil {
			t.Error("expected error for empty key, got nil")
		}
	})
}
