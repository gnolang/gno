package browser

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	boxRoundedStyle = func(r *lipgloss.Renderer) lipgloss.Style {
		b := lipgloss.RoundedBorder()
		return r.NewStyle().
			BorderStyle(b).
			Padding(0, 2)
	}

	inputStyleLeft = func(r *lipgloss.Renderer) lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return r.NewStyle().
			BorderStyle(b).
			Padding(0, 2)
	}

	infoStyle = func(r *lipgloss.Renderer) lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return boxRoundedStyle(r).BorderStyle(b)
	}
)

func (m model) View() string {
	if !m.bannerDiscarded {
		return m.bannerView()
	}

	if !m.ready {
		return "+"
	}

	mainView := fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.bodyView(), m.footerView())
	return m.zone.Scan(mainView)
}

func (m model) bannerView() string {
	banner := m.banner.View()
	if banner == "" || m.width == 0 || m.height == 0 {
		return ""
	}

	// XXX: Encapsulate banner to avoid banner glitches
	bannerView := m.render.NewStyle().Margin(1).
		Render(banner)
	widthView := m.width + 1

	return lipgloss.Place(widthView, m.height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center,
			bannerView,
			"press <enter> to continue",
		),
	)
}

func (m model) listFuncsView() string {
	return boxRoundedStyle(m.render).
		Render(m.listFuncs.View())
}

func (m model) bodyView() string {
	if m.commandInput.Focused() {
		// handle command input
		if v := m.commandInput.Value(); v != "" {
			m.listFuncs.FilterItems(v)
		} else {
			m.listFuncs.Reset()
		}

		if len(m.listFuncs.Items()) > 0 {
			m.viewport.Height = m.height - lipgloss.Height(m.listFuncsView())
		} else {
			m.viewport.Height = m.height
		}
	}

	return m.viewport.View()
}

var (
	loadingStyle = func(r *lipgloss.Renderer) lipgloss.Style {
		return r.NewStyle().
			Foreground(lipgloss.Color("#dd7878")).
			Bold(true)
	}

	navStyleEnable = func(r *lipgloss.Renderer) lipgloss.Style {
		return r.NewStyle().
			Foreground(lipgloss.Color("#fab387"))
	}

	navStyleDisable = func(r *lipgloss.Renderer) lipgloss.Style {
		return r.NewStyle().
			Foreground(lipgloss.Color("240"))
	}
)

func (m model) navView() string {
	home := navStyleEnable(m.render).Padding(0, 1).Render("[Home]")

	var style lipgloss.Style
	if m.current != nil && m.current.Prev() != nil {
		style = navStyleEnable(m.render)
	} else {
		style = navStyleDisable(m.render)
	}
	prev := style.Margin(0, 1).Render("<prev")

	if m.current != nil && m.current.Next() != nil {
		style = navStyleEnable(m.render)
	} else {
		style = navStyleDisable(m.render)
	}
	next := style.Margin(0, 1).Render("next>")

	title := m.render.NewStyle().Bold(true).Render("Gno.Land")
	if m.taskLoader.Active() {
		title = loadingStyle(m.render).Render(m.taskLoader.View())
	} else if m.devMode {
		title = lipgloss.JoinHorizontal(lipgloss.Left, title, m.connectedView())
	}

	spaceWidth := m.width / 3 // left middle and right
	return lipgloss.JoinHorizontal(lipgloss.Left,
		m.render.NewStyle().Width(spaceWidth).Padding(0, 1).
			Render(lipgloss.JoinHorizontal(lipgloss.Left,
				m.zone.Mark("prev_button", prev),
				m.zone.Mark("next_button", next),
			)),
		m.render.PlaceHorizontal(spaceWidth, lipgloss.Center, title),
		m.render.PlaceHorizontal(spaceWidth, lipgloss.Right,
			m.zone.Mark("home_button", home),
		),
	)
}

func (m model) headerView() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.navView(), m.urlView())
}

func (m model) connectedView() string {
	s := m.render.NewStyle().Bold(true)
	switch m.devClientStatus {
	case ClientStatusConnected:
		return s.Foreground(lipgloss.Color("#a6da95")).Render(" ● ")
	case ClientStatusConnecting:
		return s.Foreground(lipgloss.Color("#dd7878")).Render(" ○ ")
	case ClientStatusDisconnected:
		fallthrough
	default:
		return s.Foreground(lipgloss.Color("#eed49f")).Render(" ○ ")
	}
}

func (m model) urlView() string {
	return m.zone.Mark("url_input", boxRoundedStyle(m.render).
		Width(m.viewport.Width-2).
		Render(m.urlInput.View()))
}

func (m model) footerView() string {
	info := infoStyle(m.render).Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))

	if m.readonly {
		// On readonly, simply discard command input interface
		line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
		return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
	}

	command := m.zone.Mark("command_input", inputStyleLeft(m.render).
		Width(m.viewport.Width-lipgloss.Width(info)-5).
		Render(m.commandInput.View()))
	line := strings.Repeat("─", 3)

	powerline := lipgloss.JoinHorizontal(lipgloss.Center, command, line, info)
	if m.commandFocus && len(m.listFuncs.Items()) > 0 {
		suggestions := m.listFuncsView()
		return lipgloss.JoinVertical(lipgloss.Left, suggestions, powerline)
	}

	return powerline
}
