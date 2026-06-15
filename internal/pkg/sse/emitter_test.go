package sse

import "testing"

func TestNewEmitter(t *testing.T) {
	t.Run("returns non-nil with zero subscribers", func(t *testing.T) {
		ee := NewEmitter()
		if ee == nil {
			t.Fatal("expected non-nil emitter")
		}
		if ee.CountSubscribers() != 0 {
			t.Errorf("expected 0 subscribers, got %d", ee.CountSubscribers())
		}
	})
}

func TestEmitter_SubscribeUnsubscribe(t *testing.T) {
	t.Run("subscribe increments count", func(t *testing.T) {
		ee := NewEmitter()
		ch := make(chan *Event, 1)
		ee.Subscribe(ch)
		if ee.CountSubscribers() != 1 {
			t.Errorf("expected 1 subscriber, got %d", ee.CountSubscribers())
		}
	})

	t.Run("unsubscribe decrements count", func(t *testing.T) {
		ee := NewEmitter()
		ch := make(chan *Event, 1)
		ee.Subscribe(ch)
		ee.Unsubscribe(ch)
		if ee.CountSubscribers() != 0 {
			t.Errorf("expected 0 subscribers, got %d", ee.CountSubscribers())
		}
	})
}

func TestEmitter_RegisterEvent(t *testing.T) {
	t.Run("sends event to all subscribers", func(t *testing.T) {
		ee := NewEmitter()
		ch1 := make(chan *Event, 1)
		ch2 := make(chan *Event, 1)
		ee.Subscribe(ch1)
		ee.Subscribe(ch2)

		ee.RegisterEvent("play", "song1")

		e1 := <-ch1
		if e1.Name != "play" || e1.Data != "song1" {
			t.Errorf("ch1: expected play/song1, got %s/%s", e1.Name, e1.Data)
		}
		e2 := <-ch2
		if e2.Name != "play" || e2.Data != "song1" {
			t.Errorf("ch2: expected play/song1, got %s/%s", e2.Name, e2.Data)
		}
	})

	t.Run("unregistered channel does not receive", func(t *testing.T) {
		ee := NewEmitter()
		ch := make(chan *Event, 1)
		ee.Subscribe(ch)
		ee.Unsubscribe(ch)

		ee.RegisterEvent("pause", "")

		if len(ch) != 0 {
			t.Error("unsubscribed channel should not receive events")
		}
	})

	t.Run("no subscribers is safe", func(t *testing.T) {
		ee := NewEmitter()
		ee.RegisterEvent("play", "song1")
	})
}