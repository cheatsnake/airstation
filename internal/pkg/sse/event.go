// Package events provides a lightweight structure and utilities for creating
// and formatting Server-Sent Events (SSE) to be sent over HTTP connections.
package sse

import (
	"fmt"
	"strings"
)

// Event represents a Server-Sent Event (SSE) with a name and data payload.
type Event struct {
	Name string // The name/type of the event (used as "event:" in SSE format).
	Data string // The data payload of the event (used as "data:" in SSE format).
}

// NewEvent creates a new Event instance with the given name and data.
//
// Parameters:
//   - name: The name/type of the event.
//   - data: The string data payload associated with the event.
//
// Returns:
//   - A pointer to the newly created Event.
func NewEvent(name, data string) *Event {
	return &Event{
		Name: name,
		Data: data,
	}
}

// Stringify converts the Event into a string formatted for Server-Sent Events (SSE).
// The format includes the "event" and "data" fields as per SSE specification,
// followed by a double newline to indicate the end of the event.
//
// Returns:
//   - A string representation of the event in SSE format.
func (e *Event) Stringify() string {
	var builder strings.Builder

	if e.Name != "" {
		builder.WriteString(fmt.Sprintf("event: %s\n", e.Name))
	}

	if e.Data != "" {
		builder.WriteString(fmt.Sprintf("data: %s\n", e.Data))
	}

	builder.WriteString("\n")
	return builder.String()
}
