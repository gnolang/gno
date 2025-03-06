// modified version of ""github.com/charmbracelet/bubbles/spinner"

package browser

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Spinner struct {
	Frames []string
	FPS    time.Duration
}

// TickMsg indicates that the timer has ticked and we should render a frame.
type SpinnerTickMsg time.Time

var MeterLoader = Spinner{
	Frames: []string{
		"▱▱▱▱▱▱▱▱", "▰▱▱▱▱▱▱▱", "▰▰▱▱▱▱▱▱", "▰▰▰▱▱▱▱▱",
		"▰▰▰▰▱▱▱▱", "▰▰▰▰▰▱▱▱", "▰▰▰▰▰▰▱▱", "▰▰▰▰▰▰▰▱",
		"▰▰▰▰▰▰▰▰", "▱▰▰▰▰▰▰▰", "▱▱▰▰▰▰▰▰", "▱▱▱▰▰▰▰▰",
		"▱▱▱▱▰▰▰▰", "▱▱▱▱▱▰▰▰", "▱▱▱▱▱▱▰▰", "▱▱▱▱▱▱▱▰",
	},
	FPS: time.Second / 70, //nolint:gomnd
}

type LoaderModel struct {
	spinner Spinner
	frame   int
	task    int
}

func newLoaderModel() LoaderModel {
	return LoaderModel{
		spinner: MeterLoader,
	}
}

func (m LoaderModel) Update(msg tea.Msg) (LoaderModel, tea.Cmd) {
	switch msg.(type) {
	case SpinnerTickMsg:
		m.frame = (m.frame + 1) % len(m.spinner.Frames)
		return m, m.tick()
	default:
		return m, nil
	}
}

func (m LoaderModel) tick() tea.Cmd {
	return tea.Tick(m.spinner.FPS, func(t time.Time) tea.Msg {
		return SpinnerTickMsg(t)
	})
}

func (m LoaderModel) Tick() tea.Msg {
	return SpinnerTickMsg(time.Now())
}

func (m *LoaderModel) Active() bool {
	return m.frame > 0 || m.task > 0
}

func (m *LoaderModel) Add(i int) tea.Cmd {
	var cmd tea.Cmd
	if i > 0 {
		if m.task == 0 {
			cmd = m.Tick
		}

		m.task += i
	}
	return cmd
}

func (m *LoaderModel) Done() {
	if m.task > 0 {
		m.task -= 1
	}
}

func (m *LoaderModel) View() string {
	return m.spinner.Frames[m.frame]
}
