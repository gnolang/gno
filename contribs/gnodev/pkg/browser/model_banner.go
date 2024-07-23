package browser

import (
	"bufio"
	"fmt"
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type ModelBanner struct {
	Banner string

	enable  bool
	scanner *bufio.Scanner
	fps     time.Duration
}

func NewModelBanner(fps time.Duration, banner io.Reader) ModelBanner {
	scan := bufio.NewScanner(banner)
	return ModelBanner{
		scanner: scan,
		fps:     fps,
	}
}

func (m ModelBanner) Empty() bool {
	return m.scanner == nil
}

type tickBannerMsg struct{}

func (m ModelBanner) tick() tea.Cmd {
	return tea.Tick(m.fps, func(_ time.Time) tea.Msg {
		return tickBannerMsg{}
	})
}

func (m ModelBanner) Init() tea.Cmd {
	if m.Empty() {
		return nil
	}

	return m.tick()
}

func (m ModelBanner) Update(msg tea.Msg) (ModelBanner, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.(type) {
	case tickBannerMsg:
		if !m.Empty() && m.scanner.Scan() {
			m.Banner += fmt.Sprintln(m.scanner.Text())
			cmd = m.tick()
		}
		// XXX: handle window size
	}
	return m, cmd
}

func (m ModelBanner) View() string {
	return m.Banner
}
