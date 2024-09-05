package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
)

var Log *slog.Logger

type ColorHandler struct {
	handler slog.Handler
	w       io.Writer
	mu      sync.Mutex
}

func (h *ColorHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String()

	var color string
	switch r.Level {
	case slog.LevelDebug:
		color = "\033[36m" // Cyan
	case slog.LevelInfo:
		color = "\033[32m" // Green
	case slog.LevelWarn:
		color = "\033[33m" // Yellow
	case slog.LevelError:
		color = "\033[31m" // Red
	default:
		color = "\033[0m" // Default
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Format time
	timeStr := r.Time.Format("2006-01-02 15:04:05")

	// Format attributes
	attrs := make(map[string]interface{})
	r.Attrs(
		func(a slog.Attr) bool {
			attrs[a.Key] = a.Value.Any()
			return true
		},
	)
	attrsJSON, _ := json.Marshal(attrs)

	// Print colored output
	fmt.Fprintf(h.w, "%s%s [%s] %s %s\033[0m\n", color, timeStr, level, r.Message, string(attrsJSON))

	return nil
}

func (h *ColorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ColorHandler{handler: h.handler.WithAttrs(attrs), w: h.w}
}

func (h *ColorHandler) WithGroup(name string) slog.Handler {
	return &ColorHandler{handler: h.handler.WithGroup(name), w: h.w}
}

func (h *ColorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func NewColorHandler(w io.Writer, opts *slog.HandlerOptions) *ColorHandler {
	return &ColorHandler{
		handler: slog.NewJSONHandler(w, opts),
		w:       w,
	}
}

func Init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	handler := NewColorHandler(os.Stdout, opts)
	Log = slog.New(handler)

	slog.SetDefault(Log)
}
