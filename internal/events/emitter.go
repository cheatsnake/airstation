package events

import (
	"sync"
)

// Emitter manages a set of subscribers and broadcasts events to them.
type Emitter struct {
	subscribers sync.Map // A thread-safe map storing subscriber channels.
}

// NewEmitter creates and returns a new Emitter instance.
//
// Returns:
//   - A pointer to a new Emitter.
func NewEmitter() *Emitter {
	return &Emitter{}
}

// RegisterEvent broadcasts an event with the specified name and data
// to all currently subscribed channels.
//
// Parameters:
//   - name: The event name/type.
//   - data: The string payload of the event.
func (ee *Emitter) RegisterEvent(name, data string) {
	event := NewEvent(name, data)
	ee.subscribers.Range(func(key, value any) bool {
		eventChan := key.(chan *Event)
		eventChan <- event
		return true
	})
}

// CountSubscribers returns the number of currently active subscribers.
//
// Returns:
//   - An integer count of subscriber channels.
func (ee *Emitter) CountSubscribers() int {
	count := 0
	ee.subscribers.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}

// Subscribe adds a new subscriber channel to receive events.
//
// Parameters:
//   - eventChan: A channel to which events will be sent.
func (ee *Emitter) Subscribe(eventChan chan *Event) {
	ee.subscribers.Store(eventChan, true)
}

// Unsubscribe removes a previously added subscriber channel.
//
// Parameters:
//   - eventChan: The channel to remove from the list of subscribers.
func (ee *Emitter) Unsubscribe(eventChan chan *Event) {
	ee.subscribers.Delete(eventChan)
}
