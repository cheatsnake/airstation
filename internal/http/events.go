package http

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type serverSideEvent struct {
	Name string
	Data string
}

func (s *Server) eventCountConnections() {
	count := 0

	s.connections.Range(func(key, value any) bool {
		count++
		return true
	})

	event := serverSideEvent{
		Name: "count_connections",
		Data: strconv.Itoa(count),
	}

	s.connections.Range(func(key, value any) bool {
		eventChan := key.(chan serverSideEvent)
		eventChan <- event
		return true
	})
}

func (e *serverSideEvent) Stringify() string {
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

func (s *Server) runIntervalEvents() {
	connectionCountTicker := time.Tick(5 * time.Second)

	go func() {
		for range connectionCountTicker {
			s.eventCountConnections()
		}
	}()
}
