package debuglogger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	yellow = "\033[33m"
	green  = "\033[32m"
	blue   = "\033[34m"
	gray   = "\033[90m"
)

type DevHandler struct {
	w     io.Writer
	opts  slog.HandlerOptions
	mu    *sync.Mutex
	attrs []slog.Attr
}

func NewDevelopmentHandler(w io.Writer, opts *slog.HandlerOptions) *DevHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &DevHandler{
		w:    w,
		opts: *opts,
		mu:   &sync.Mutex{},
	}
}

func (h *DevHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

func (h *DevHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var buf []byte
	color := h.levelColor(r.Level)
	if !r.Time.IsZero() {
		t := r.Time.Local().Format("15:04:05")
		buf = append(buf, gray...)
		buf = append(buf, t...)
		buf = append(buf, ' ')
	}
	buf = append(buf, color...)
	buf = append(buf, '[')
	buf = append(buf, r.Level.String()...)
	buf = append(buf, ']', ' ')
	buf = append(buf, reset...)
	buf = append(buf, r.Message...)
	attrs := make([]slog.Attr, 0, len(h.attrs)+r.NumAttrs())
	attrs = append(attrs, h.attrs...)
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	if len(attrs) > 0 {
		buf = append(buf, ' ')
		for i, attr := range attrs {
			if i > 0 {
				buf = append(buf, ' ')
			}
			buf = append(buf, h.attrColor(attr.Key)...)
			buf = append(buf, attr.Key...)
			buf = append(buf, '=')
			buf = append(buf, reset...)
			buf = append(buf, h.formatValue(attr.Value)...)
		}
	}
	if r.Level >= slog.LevelError {
		r.Attrs(func(a slog.Attr) bool {
			if a.Key == "error" {
				buf = append(buf, '\n', '\t')
				buf = append(buf, red...)
				buf = append(buf, "error: "...)
				buf = append(buf, a.Value.String()...)
				buf = append(buf, reset...)
			}
			return true
		})
	}

	buf = append(buf, '\n')
	_, err := h.w.Write(buf)
	return err
}

func (h *DevHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &DevHandler{
		w:     h.w,
		opts:  h.opts,
		mu:    h.mu,
		attrs: append(h.attrs, attrs...),
	}
}

func (h *DevHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *DevHandler) levelColor(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return red
	case level >= slog.LevelWarn:
		return yellow
	case level >= slog.LevelInfo:
		return green
	default: 
		return blue
	}
}

func (h *DevHandler) attrColor(key string) string {
	switch key {
	case "method", "path", "status":
		return blue
	case "duration", "operation", "user_id":
		return green
	default:
		return ""
	}
}

func (h *DevHandler) formatValue(v slog.Value) string {
	switch v.Kind() {
	case slog.KindInt64:
		return fmt.Sprintf("%d", v.Int64())
	case slog.KindUint64:
		return fmt.Sprintf("%d", v.Uint64())
	case slog.KindFloat64:
		return fmt.Sprintf("%.2f", v.Float64())
	case slog.KindBool:
		return fmt.Sprintf("%t", v.Bool())
	case slog.KindDuration:
		return v.Duration().Round(time.Microsecond).String()
	case slog.KindTime:
		return v.Time().Format("15:04:05")
	default:
		return v.String()
	}
}
