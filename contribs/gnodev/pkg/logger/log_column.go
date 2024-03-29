package logger

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
)

func NewColumnLogger(w io.Writer, level slog.Level, profile termenv.Profile) *slog.Logger {
	charmLogger := log.NewWithOptions(w, log.Options{
		ReportTimestamp: false,
		ReportCaller:    false,
		Prefix:          "",
	})

	// Default column output
	charmLogger.SetOutput(&columnWriter{
		style:  lipgloss.NewStyle(),
		prefix: "",
		writer: w,
	})

	columnHandler := &columnLogger{
		Logger: charmLogger,
		writer: w,
		prefix: charmLogger.GetPrefix(),
	}

	charmLogger.SetOutput(newColumeWriter(lipgloss.NewStyle(), "", w))
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

	return slog.New(columnHandler)
}

type columnLogger struct {
	*log.Logger

	prefix       string
	writer       io.Writer
	colorProfile termenv.Profile
}

func (cl *columnLogger) WithGroup(name string) slog.Handler {
	if cl.prefix != "" {
		name = fmt.Sprintf("%.1s.%s", cl.prefix, name)
	}

	fg := ColorFromString(name, 0.6, 0.4)
	baseStyle := lipgloss.NewStyle().Foreground(fg)

	styles := defaultStyles()
	styles.Message = styles.Message.Foreground(fg)

	nlog := cl.Logger.With() // clone logger
	nlog.SetOutput(newColumeWriter(baseStyle, name, cl.writer))
	nlog.SetColorProfile(cl.colorProfile)
	nlog.SetStyles(styles)
	return &columnLogger{
		Logger: nlog,
		prefix: name,
		writer: cl.writer,
	}
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
			if _, err = cl.writer.Write([]byte(lf)); err != nil {
				return n, err
			}
			n++
			buf = buf[1:]
		}
	}

	return n, nil
}

// defaultStyles returns the default styles.
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
