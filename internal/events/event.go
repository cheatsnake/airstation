package events

import (
	"fmt"
	"strings"
)

type Event struct {
	Name string
	Data string
}

func NewEvent(name, data string) *Event {
	return &Event{
		Name: name,
		Data: data,
	}
}

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
