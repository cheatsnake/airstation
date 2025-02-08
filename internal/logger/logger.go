package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"
)

type customHandler struct {
	output io.Writer
	opts   *slog.HandlerOptions
	group  string
}

func (h *customHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if h.opts != nil && h.opts.Level != nil {
		return level >= h.opts.Level.Level()
	}
	return level >= slog.LevelInfo
}

func (h *customHandler) Handle(ctx context.Context, r slog.Record) error {
	logMeta := r.Time.Format(time.DateTime) + " " + r.Level.String() + " "
	if h.group != "" {
		logMeta += h.group + ": "
	}

	_, err := h.output.Write([]byte(logMeta + r.Message + " "))
	if err != nil {
		return err
	}

	r.Attrs(func(attr slog.Attr) bool {
		println(attr.Key)
		_, err := h.output.Write([]byte(attr.Value.String() + " "))
		return err == nil
	})

	_, err = h.output.Write([]byte("\n"))
	return err
}

func (h *customHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := &customHandler{output: h.output, opts: h.opts, group: h.group}
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "", 0)
	for _, attr := range attrs {
		r.AddAttrs(attr)
	}
	return newHandler
}

func (h *customHandler) WithGroup(name string) slog.Handler {
	return &customHandler{output: h.output, opts: h.opts, group: name}
}

func newCustomHandler(w io.Writer, opts *slog.HandlerOptions) *customHandler {
	return &customHandler{output: w, opts: opts}
}

func New() *slog.Logger {
	handler := newCustomHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := slog.New(handler)
	return logger
}
