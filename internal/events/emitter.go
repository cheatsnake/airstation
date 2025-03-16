package events

import (
	"sync"
)

type Emitter struct {
	subscribers sync.Map
}

func NewEmitter() *Emitter {
	return &Emitter{}
}

func (ee *Emitter) RegisterEvent(name, data string) {
	event := NewEvent(name, data)
	ee.subscribers.Range(func(key, value any) bool {
		eventChan := key.(chan *Event)
		eventChan <- event
		return true
	})
}

func (ee *Emitter) CountSubscribers() int {
	count := 0
	ee.subscribers.Range(func(key, value any) bool {
		count++
		return true
	})

	return count
}

func (ee *Emitter) Subscribe(eventChan chan *Event) {
	ee.subscribers.Store(eventChan, true)
}

func (ee *Emitter) Unsubscribe(eventChan chan *Event) {
	ee.subscribers.Delete(eventChan)
}
