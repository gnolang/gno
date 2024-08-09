package browser

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type ModelBanner struct {
	Banner string

	offset     int
	frameIndex int
	frames     [][]string

	fps time.Duration
}

func NewModelBanner(fps time.Duration, frames []string) ModelBanner {
	splited := make([][]string, len(frames))
	for i, frame := range frames {
		lines := strings.Split(frame, "\n")
		for j, line := range lines {
			lines[j] = line + "\033[0m"
		}
		splited[i] = lines
	}

	return ModelBanner{
		frames: splited,
		fps:    fps,
	}
}

func (m ModelBanner) Empty() bool {
	return m.frames == nil
}

type (
	tickBannerMsg       struct{}
	tickBannerOffsetMsg struct{}
)

func (m ModelBanner) tick() tea.Cmd {
	return tea.Tick(m.fps, func(_ time.Time) tea.Msg {
		return tickBannerMsg{}
	})
}

func (m ModelBanner) tickOffset() tea.Cmd {
	return tea.Tick(time.Second/10, func(_ time.Time) tea.Msg {
		return tickBannerOffsetMsg{}
	})
}

func (m ModelBanner) Init() tea.Cmd {
	if m.Empty() {
		return nil
	}

	return tea.Batch(m.tickOffset(), m.tick())
}

func (m ModelBanner) Update(msg tea.Msg) (ModelBanner, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.(type) {
	case tickBannerOffsetMsg:
		frame := m.frames[m.frameIndex]
		m.Banner = getFrameLinesOffset(frame, m.offset)
		if m.offset < (len(frame) / 2) {
			m.offset++
			cmd = m.tickOffset()
		}

	case tickBannerMsg:
		frame := m.frames[m.frameIndex]
		m.Banner = getFrameLinesOffset(frame, m.offset)
		m.frameIndex = (m.frameIndex + 1) % len(m.frames) // move to next frame
		cmd = m.tick()
		// XXX: handle window size
	}
	return m, cmd
}

func (m ModelBanner) View() string {
	return m.Banner
}

func getFrameLinesOffset(lines []string, offset int) string {
	middle := len(lines) / 2
	if offset < middle {
		start := middle - min(middle, offset)
		end := middle + min(middle, offset)
		lines = lines[start:end]
	}

	return strings.Join(lines, "\n")
}
