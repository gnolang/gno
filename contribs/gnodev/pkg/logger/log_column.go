package logger

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"io"
	"log/slog"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
)

func NewColumnLogger(w io.Writer, level slog.Level, color bool) *slog.Logger {
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

	colorProfile := termenv.TrueColor
	if !color {
		colorProfile = termenv.Ascii
	}

	columnHandler := &columnLogger{
		Logger:       charmLogger,
		writer:       w,
		prefix:       charmLogger.GetPrefix(),
		colorProfile: colorProfile,
	}

	charmLogger.SetStyles(DefaultStyles())
	charmLogger.SetColorProfile(colorProfile)
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

	nlog := cl.Logger.With() // clone logger
	baseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(strconv.Itoa(stringToColor(name))))
	nlog.SetOutput(NewColumeWriter(baseStyle, name, cl.writer))
	nlog.SetColorProfile(cl.colorProfile)
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

func NewColumeWriter(baseStyle lipgloss.Style, prefix string, writer io.Writer) *columnWriter {
	style := baseStyle.
		Border(lipgloss.ThickBorder(), false, true, false, false).
		BorderForeground(baseStyle.GetForeground()).
		Bold(true).
		Width(12)
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

func stringToColor(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int((h.Sum32()+10)%255) + 1
}
