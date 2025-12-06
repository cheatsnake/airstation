package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

type logEntry struct {
	Time    string         `json:"time"`
	Level   string         `json:"level"`
	Group   string         `json:"group,omitempty"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type customJSONHandler struct {
	output io.Writer
	opts   *slog.HandlerOptions
	group  string
	attrs  []slog.Attr
}

func (h *customJSONHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if h.opts != nil && h.opts.Level != nil {
		return level >= h.opts.Level.Level()
	}
	return level >= slog.LevelInfo
}

func (h *customJSONHandler) Handle(ctx context.Context, r slog.Record) error {
	details := make(map[string]any)

	for _, attr := range h.attrs {
		details[attr.Key] = attr.Value.Any()
	}

	r.Attrs(func(attr slog.Attr) bool {
		details[attr.Key] = attr.Value.Any()
		return true
	})

	entry := logEntry{
		Time:    r.Time.Format(time.DateTime),
		Level:   r.Level.String(),
		Message: r.Message,
		Group:   h.group,
	}

	if len(details) > 0 {
		entry.Details = details
	}

	var buf bytes.Buffer
	buf.WriteString("{")
	fmt.Fprintf(&buf, `"time":"%s"`, entry.Time)
	fmt.Fprintf(&buf, `,"level":"%s"`, entry.Level)
	if entry.Group != "" {
		fmt.Fprintf(&buf, `,"group":"%s"`, entry.Group)
	}
	fmt.Fprintf(&buf, `,"message":"%s"`, jsonEscape(entry.Message))

	if len(details) > 0 {
		buf.WriteString(`,"details":`)
		detailsJSON, err := json.Marshal(details)
		if err != nil {
			return err
		}
		buf.Write(detailsJSON)
	}

	buf.WriteString("}\n")

	_, err := h.output.Write(buf.Bytes())
	return err
}

func jsonEscape(s string) string {
	b, _ := json.Marshal(s)
	return string(b[1 : len(b)-1])
}

func (h *customJSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &customJSONHandler{
		output: h.output,
		opts:   h.opts,
		group:  h.group,
		attrs:  newAttrs,
	}
}

func (h *customJSONHandler) WithGroup(name string) slog.Handler {
	return &customJSONHandler{
		output: h.output,
		opts:   h.opts,
		group:  name,
		attrs:  h.attrs,
	}
}

func newCustomJSONHandler(w io.Writer, opts *slog.HandlerOptions) *customJSONHandler {
	return &customJSONHandler{
		output: w,
		opts:   opts,
		attrs:  []slog.Attr{},
	}
}

func New() *slog.Logger {
	handler := newCustomJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	return slog.New(handler)
}
