// Package termlog renders slog records as colored, human-readable terminal lines.
package termlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"sync"
)

// ---- ANSI color codes

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// Handler is a slog.Handler that produces colored, human-readable output.
//
// Non-verbose (Info+): focuses on scenario progress — clean, minimal.
// Verbose (Debug): includes component tags and structured attrs.
type Handler struct {
	w       io.Writer
	level   slog.Level
	attrs   []slog.Attr // pre-resolved group attrs (component, etc.)
	mu      *sync.Mutex
	verbose bool
}

func NewHandler(w io.Writer, verbose bool) *Handler {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	return &Handler{
		w:       w,
		level:   level,
		mu:      &sync.Mutex{},
		verbose: verbose,
	}
}

func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Level color
	var levelStr string
	switch {
	case r.Level >= slog.LevelError:
		levelStr = colorRed + "ERR" + colorReset
	case r.Level >= slog.LevelWarn:
		levelStr = colorYellow + "WRN" + colorReset
	case r.Level >= slog.LevelInfo:
		levelStr = colorGreen + "INF" + colorReset
	default:
		levelStr = colorGray + "DBG" + colorReset
	}

	// Component from pre-resolved attrs
	component := ""
	for _, a := range h.attrs {
		if a.Key == "component" {
			component = a.Value.String()
		}
	}

	// Build the line
	if h.verbose {
		// Verbose: [level] [component] message key=value ...
		comp := ""
		if component != "" {
			comp = colorCyan + "[" + component + "]" + colorReset + " "
		}
		var line strings.Builder
		line.WriteString(fmt.Sprintf("%s %s%s", levelStr, comp, r.Message))

		// Append attrs from pre-resolved + record
		r.Attrs(func(a slog.Attr) bool {
			line.WriteString(attrText(a))
			return true
		})
		for _, a := range h.attrs {
			if a.Key == "component" {
				continue
			}
			line.WriteString(attrText(a))
		}

		fmt.Fprintln(h.w, line.String())
	} else {
		// Non-verbose: clean output focused on user-relevant info.
		// Component prefix for context, then message + key attrs inline.
		prefix := ""
		if component != "" {
			prefix = colorBold + component + colorReset + ": "
		}
		line := fmt.Sprintf("%s %s%s", levelStr, prefix, r.Message)

		r.Attrs(func(a slog.Attr) bool {
			line += attrText(a)
			return true
		})

		fmt.Fprintln(h.w, line)
	}

	return nil
}

// attrText renders one attribute, resolving a value that computes itself:
// slog.Value.String on an unresolved LogValuer prints the thing that would have
// computed the value rather than the value.
func attrText(a slog.Attr) string {
	return " " + colorGray + a.Key + "=" + colorReset + a.Value.Resolve().String()
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		w:       h.w,
		level:   h.level,
		attrs:   append(slices.Clone(h.attrs), attrs...),
		mu:      h.mu,
		verbose: h.verbose,
	}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	// slog requires an empty name to be a no-op. Left to the branch below it
	// would label every later record with a group that was never opened.
	if name == "" {
		return h
	}
	// Not used in our codebase; treat as attr prefix.
	return &Handler{
		w:       h.w,
		level:   h.level,
		attrs:   append(slices.Clone(h.attrs), slog.String("group", name)),
		mu:      h.mu,
		verbose: h.verbose,
	}
}
