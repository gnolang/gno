package logger

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
)

func NewColumnLogger(w io.Writer, level slog.Level, profile termenv.Profile) *ColumnLogger {
	charmLogger := log.NewWithOptions(w, log.Options{
		ReportTimestamp: false,
		ReportCaller:    false,
		Prefix:          "",
	})

	// Default column output
	defaultOutput := newColumeWriter(lipgloss.NewStyle(), "", w)
	charmLogger.SetOutput(defaultOutput)
	charmLogger.SetStyles(defaultStyles())
	charmLogger.SetColorProfile(profile)
	charmLogger.SetReportCaller(false)
	switch level {
	case slog.LevelDebug:
		charmLogger.SetLevel(log.DebugLevel)
	case slog.LevelError:
		charmLogger.SetLevel(log.ErrorLevel)
	case slog.LevelInfo:
		charmLogger.SetLevel(log.InfoLevel)
	case slog.LevelWarn:
		charmLogger.SetLevel(log.WarnLevel)
	default:
		panic("invalid slog level")
	}

	return &ColumnLogger{
		Logger: charmLogger,
		writer: w,
		prefix: charmLogger.GetPrefix(),
		colors: map[string]lipgloss.Color{},
	}
}

type ColumnLogger struct {
	*log.Logger

	prefix       string
	writer       io.Writer
	colorProfile termenv.Profile

	colors   map[string]lipgloss.Color
	muColors sync.RWMutex
}

func (cl *ColumnLogger) WithGroup(group string) slog.Handler {
	cl.muColors.RLock()
	defer cl.muColors.RUnlock()

	if cl.prefix != "" {
		group = fmt.Sprintf("%.1s.%s", cl.prefix, group)
	}

	// check if we already know this group
	fg, ok := cl.colors[group]
	if !ok {
		// generate bright color based on the group name
		fg = colorFromString(group, 0.5, 0.6)
	}
	baseStyle := lipgloss.NewStyle().Foreground(fg)

	nlog := cl.Logger.With() // clone logger
	nlog.SetOutput(newColumeWriter(baseStyle, group, cl.writer))
	nlog.SetColorProfile(cl.colorProfile)
	return &ColumnLogger{
		Logger: nlog,
		prefix: group,
		writer: cl.writer,
	}
}

func (cl *ColumnLogger) RegisterGroupColor(group string, color lipgloss.Color) {
	cl.muColors.Lock()
	cl.colors[group] = color
	cl.muColors.Unlock()
}

var lf = []byte{'\n'}

type columnWriter struct {
	inline bool
	style  lipgloss.Style
	prefix string
	writer io.Writer
}

func newColumeWriter(baseStyle lipgloss.Style, prefix string, writer io.Writer) *columnWriter {
	const width = 12

	style := baseStyle.
		Border(lipgloss.ThickBorder(), false, true, false, false).
		BorderForeground(baseStyle.GetForeground()).
		Bold(true).
		Width(width)

	if len(prefix) >= width {
		prefix = prefix[:width-3] + "..."
	}

	return &columnWriter{style: style, prefix: prefix, writer: writer}
}

func (cl *columnWriter) Write(buf []byte) (n int, err error) {
	for line := 0; len(buf) > 0; line++ {
		i := bytes.IndexByte(buf, '\n')
		todo := len(buf)
		if i >= 0 {
			todo = i
		}

		if !cl.inline {
			var prefix string
			if line == 0 {
				prefix = cl.prefix
			}

			fmt.Fprint(cl.writer, cl.style.Render(prefix)+" ")
		}

		var nn int
		nn, err = cl.writer.Write(buf[:todo])
		n += nn
		if err != nil {
			return n, err
		}
		buf = buf[todo:]

		if cl.inline = i < 0; !cl.inline {
			if _, err = cl.writer.Write(lf); err != nil {
				return n, err
			}
			n++
			buf = buf[1:]
		}
	}

	return n, nil
}

// defaultStyles returns the default lipgloss styles for column logger
func defaultStyles() *log.Styles {
	style := log.DefaultStyles()
	style.Levels = map[log.Level]lipgloss.Style{
		log.DebugLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.DebugLevel.String())).
			Bold(true).
			MaxWidth(1).
			Foreground(lipgloss.Color("63")),
		log.InfoLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.InfoLevel.String())).
			MaxWidth(1).
			Foreground(lipgloss.Color("12")),

		log.WarnLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.WarnLevel.String())).
			Bold(true).
			MaxWidth(1).
			Foreground(lipgloss.Color("192")),
		log.ErrorLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.ErrorLevel.String())).
			Bold(true).
			MaxWidth(1).
			Foreground(lipgloss.Color("204")),
		log.FatalLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.FatalLevel.String())).
			Bold(true).
			MaxWidth(1).
			Foreground(lipgloss.Color("134")),
	}
	style.Keys = map[string]lipgloss.Style{
		"err": lipgloss.NewStyle().
			Foreground(lipgloss.Color("204")),
		"error": lipgloss.NewStyle().
			Foreground(lipgloss.Color("204")),
	}
	style.Values = map[string]lipgloss.Style{
		"err": lipgloss.NewStyle().
			Foreground(lipgloss.Color("204")),
		"error": lipgloss.NewStyle().
			Foreground(lipgloss.Color("204")),
	}

	return style
}
