package sqlite

import (
	"testing"

	"github.com/cheatsnake/airstation/internal/track"
)

func TestQueueStore_AddAndRetrieve(t *testing.T) {
	inst := setupTestDB(t)
	a := addTestTrack(t, inst, "Track A", "/a.aac", 60.0, 128)
	b := addTestTrack(t, inst, "Track B", "/b.aac", 120.0, 192)
	c := addTestTrack(t, inst, "Track C", "/c.aac", 180.0, 256)

	err := inst.QueueStore.AddToQueue([]*track.Track{a, b, c})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	queue, err := inst.QueueStore.Queue()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(queue) != 3 {
		t.Errorf("expected 3 tracks in queue, got %d", len(queue))
	}
	if queue[0].ID != a.ID {
		t.Errorf("expected first track %q, got %q", a.ID, queue[0].ID)
	}
}

func TestQueueStore_AddToQueue_Deduplicate(t *testing.T) {
	inst := setupTestDB(t)
	a := addTestTrack(t, inst, "Track A", "/a.aac", 60.0, 128)
	b := addTestTrack(t, inst, "Track B", "/b.aac", 120.0, 192)

	err := inst.QueueStore.AddToQueue([]*track.Track{a})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = inst.QueueStore.AddToQueue([]*track.Track{a, b})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	queue, err := inst.QueueStore.Queue()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(queue) != 2 {
		t.Errorf("expected 2 tracks after dedup, got %d", len(queue))
	}
}

func TestQueueStore_ReorderQueue(t *testing.T) {
	inst := setupTestDB(t)
	a := addTestTrack(t, inst, "Track A", "/a.aac", 60.0, 128)
	b := addTestTrack(t, inst, "Track B", "/b.aac", 120.0, 192)
	c := addTestTrack(t, inst, "Track C", "/c.aac", 180.0, 256)

	inst.QueueStore.AddToQueue([]*track.Track{a, b, c})

	err := inst.QueueStore.ReorderQueue([]string{c.ID, a.ID, b.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	queue, err := inst.QueueStore.Queue()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(queue) != 3 {
		t.Fatalf("expected 3 tracks, got %d", len(queue))
	}
	if queue[0].ID != c.ID {
		t.Errorf("expected first track %q after reorder, got %q", c.ID, queue[0].ID)
	}
	if queue[1].ID != a.ID {
		t.Errorf("expected second track %q after reorder, got %q", a.ID, queue[1].ID)
	}
}

func TestQueueStore_RemoveFromQueue(t *testing.T) {
	inst := setupTestDB(t)
	a := addTestTrack(t, inst, "Track A", "/a.aac", 60.0, 128)
	b := addTestTrack(t, inst, "Track B", "/b.aac", 120.0, 192)

	inst.QueueStore.AddToQueue([]*track.Track{a, b})

	err := inst.QueueStore.RemoveFromQueue([]string{a.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	queue, err := inst.QueueStore.Queue()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(queue) != 1 {
		t.Errorf("expected 1 track, got %d", len(queue))
	}
	if queue[0].ID != b.ID {
		t.Errorf("expected track B remaining, got %q", queue[0].ID)
	}
}

func TestQueueStore_SpinQueue(t *testing.T) {
	inst := setupTestDB(t)
	a := addTestTrack(t, inst, "Track A", "/a.aac", 60.0, 128)
	b := addTestTrack(t, inst, "Track B", "/b.aac", 120.0, 192)
	c := addTestTrack(t, inst, "Track C", "/c.aac", 180.0, 256)

	inst.QueueStore.AddToQueue([]*track.Track{a, b, c})

	err := inst.QueueStore.SpinQueue()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	queue, err := inst.QueueStore.Queue()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(queue) != 3 {
		t.Fatalf("expected 3 tracks, got %d", len(queue))
	}
	if queue[0].ID != b.ID {
		t.Errorf("expected track B first after spin, got %q", queue[0].ID)
	}
	if queue[2].ID != a.ID {
		t.Errorf("expected track A last after spin, got %q", queue[2].ID)
	}
}

func TestQueueStore_SpinQueue_Empty(t *testing.T) {
	inst := setupTestDB(t)

	err := inst.QueueStore.SpinQueue()
	if err != nil {
		t.Fatalf("unexpected error on empty queue: %v", err)
	}
}

func TestQueueStore_CurrentAndNextTrack(t *testing.T) {
	inst := setupTestDB(t)
	a := addTestTrack(t, inst, "Track A", "/a.aac", 60.0, 128)
	b := addTestTrack(t, inst, "Track B", "/b.aac", 120.0, 192)

	t.Run("empty queue returns nil", func(t *testing.T) {
		current, next, err := inst.QueueStore.CurrentAndNextTrack()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if current != nil || next != nil {
			t.Errorf("expected nil tracks for empty queue")
		}
	})

	t.Run("single track returns it as both current and next", func(t *testing.T) {
		inst.QueueStore.AddToQueue([]*track.Track{a})
		current, next, err := inst.QueueStore.CurrentAndNextTrack()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if current.ID != a.ID {
			t.Errorf("expected current %q, got %q", a.ID, current.ID)
		}
		if next.ID != a.ID {
			t.Errorf("expected next same as current when single, got %q", next.ID)
		}
		inst.QueueStore.RemoveFromQueue([]string{a.ID})
	})

	t.Run("two tracks return distinct current and next", func(t *testing.T) {
		inst.QueueStore.AddToQueue([]*track.Track{a, b})
		current, next, err := inst.QueueStore.CurrentAndNextTrack()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if current.ID != a.ID {
			t.Errorf("expected current %q, got %q", a.ID, current.ID)
		}
		if next.ID != b.ID {
			t.Errorf("expected next %q, got %q", b.ID, next.ID)
		}
	})
}
