// Package logger wires slog to the right handler per environment:
//   - dev:  colored, column-aligned, human-first console output
//   - prod: structured JSON for log aggregators (Loki, Datadog, CloudWatch...)
//
// Application code always calls slog.* directly; only the handler swaps.
package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// New returns the logger for the environment.
func New(level string, production bool) *slog.Logger {
	lvl := parseLevel(level)
	if production {
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
	}
	return slog.New(newPrettyHandler(os.Stdout, lvl))
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// ---------- ANSI helpers ----------

const (
	ansiReset   = "\033[0m"
	ansiBold    = "\033[1m"
	ansiDim     = "\033[2m"
	ansiRed     = "\033[31m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiBlue    = "\033[34m"
	ansiMagenta = "\033[35m"
	ansiCyan    = "\033[36m"
)

func paint(color bool, code, s string) string {
	if !color || s == "" {
		return s
	}
	return code + s + ansiReset
}

// colorSupported: colors only on a real terminal, and honor NO_COLOR.
// Piping output (ftpl run dev | tee log.txt) automatically gets plain text.
func colorSupported() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// ---------- pretty console handler (dev) ----------

type prettyHandler struct {
	w     io.Writer
	mu    *sync.Mutex
	level slog.Level
	color bool
	attrs []slog.Attr
}

func newPrettyHandler(w io.Writer, level slog.Level) *prettyHandler {
	return &prettyHandler{w: w, mu: &sync.Mutex{}, level: level, color: colorSupported()}
}

func (h *prettyHandler) Enabled(_ context.Context, l slog.Level) bool { return l >= h.level }

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	n := *h
	n.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &n
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	n := *h
	n.attrs = append(append([]slog.Attr{}, slog.String("group", name)), h.attrs...)
	return &n
}

func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	sep := paint(h.color, ansiDim, " │ ")
	var sb strings.Builder

	// Column 1: short clock, dim.
	sb.WriteString(paint(h.color, ansiDim, r.Time.Format("15:04:05")))
	sb.WriteString(sep)

	// Column 2: fixed-width level badge.
	sb.WriteString(levelBadge(h.color, r.Level))
	sb.WriteString(sep)

	// Column 3: message, bold.
	sb.WriteString(paint(h.color, ansiBold, r.Message))

	// Columns 4+: attrs, colored per key, empties skipped.
	for _, a := range h.attrs {
		if s := h.formatAttr(a); s != "" {
			sb.WriteString("  ")
			sb.WriteString(s)
		}
	}
	r.Attrs(func(a slog.Attr) bool {
		if s := h.formatAttr(a); s != "" {
			sb.WriteString("  ")
			sb.WriteString(s)
		}
		return true
	})

	sb.WriteString("\n")

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.w, sb.String())
	return err
}

// formatAttr renders one key=value with per-key semantics:
// status colored by class, latency colored by speed, errors red,
// empty values omitted entirely (kills the visual noise).
func (h *prettyHandler) formatAttr(a slog.Attr) string {
	a.Value = a.Value.Resolve()
	val := valueString(a.Value)
	if val == "" {
		return ""
	}
	key := paint(h.color, ansiDim, a.Key+"=")

	switch a.Key {
	case "status":
		return key + statusColor(h.color, val)
	case "latency":
		return key + latencyColor(h.color, val)
	case "method":
		return key + paint(h.color, ansiBlue+ansiBold, val)
	case "path":
		return key + paint(h.color, ansiCyan, val)
	case "error", "err":
		return key + paint(h.color, ansiRed+ansiBold, val)
	default:
		return key + val
	}
}

func levelBadge(color bool, l slog.Level) string {
	var tag, col string
	switch {
	case l >= slog.LevelError:
		tag, col = "ERROR", ansiRed+ansiBold
	case l >= slog.LevelWarn:
		tag, col = "WARN ", ansiYellow+ansiBold
	case l <= slog.LevelDebug:
		tag, col = "DEBUG", ansiDim
	default:
		tag, col = "INFO ", ansiGreen+ansiBold
	}
	return paint(color, col, tag)
}

// statusColor: 2xx green, 3xx cyan, 4xx yellow, 5xx red+bold.
func statusColor(color bool, val string) string {
	code, err := strconv.Atoi(val)
	if err != nil {
		return val
	}
	switch {
	case code >= 500:
		return paint(color, ansiRed+ansiBold, val)
	case code >= 400:
		return paint(color, ansiYellow, val)
	case code >= 300:
		return paint(color, ansiCyan, val)
	default:
		return paint(color, ansiGreen, val)
	}
}

// latencyColor: instant magenta, >300ms yellow, >1s red+bold —
// slow requests jump out of the scroll without grepping anything.
func latencyColor(color bool, val string) string {
	d, err := time.ParseDuration(val)
	if err != nil {
		return paint(color, ansiMagenta, val)
	}
	switch {
	case d > time.Second:
		return paint(color, ansiRed+ansiBold, val)
	case d > 300*time.Millisecond:
		return paint(color, ansiYellow+ansiBold, val)
	default:
		return paint(color, ansiMagenta, val)
	}
}

func valueString(v slog.Value) string {
	switch v.Kind() {
	case slog.KindString:
		return v.String()
	case slog.KindDuration:
		return v.Duration().String()
	case slog.KindInt64:
		return strconv.FormatInt(v.Int64(), 10)
	case slog.KindUint64:
		return strconv.FormatUint(v.Uint64(), 10)
	case slog.KindBool:
		return strconv.FormatBool(v.Bool())
	case slog.KindFloat64:
		return strconv.FormatFloat(v.Float64(), 'f', -1, 64)
	default:
		b, err := json.Marshal(v.Any())
		if err != nil {
			return fmt.Sprint(v.Any())
		}
		return string(b)
	}
}
