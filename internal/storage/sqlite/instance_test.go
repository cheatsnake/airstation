package sqlite

import (
	"io"
	"log/slog"
	"testing"
)

func TestInstance_New(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	inst, err := New(":memory:", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inst == nil {
		t.Fatal("expected non-nil instance")
	}
	inst.Close()
}

func TestInstance_Close(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	inst, err := New(":memory:", log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = inst.Close()
	if err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}
}
