package main

import (
	"bytes"
	clist "container/list"
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/muesli/reflow/wordwrap"
)

// var noopLogger = log.NewNopLogger()

const gnoPrefix = "gno.land/"

// You generally won't need this unless you're processing stuff with
// complicated ANSI escape sequences. Turn it on if you notice flickering.
//
// Also keep in mind that high performance rendering only works for programs
// that use the full size of the terminal. We're enabling that below with
// tea.EnterAltScreen().
const useHighPerformanceRenderer = false

var (
	navStyleEnable = func(r *lipgloss.Renderer) lipgloss.Style {
		return r.NewStyle().
			Foreground(lipgloss.Color("#fab387"))
	}

	navStyleDisable = func(r *lipgloss.Renderer) lipgloss.Style {
		return r.NewStyle().
			Foreground(lipgloss.Color("240"))
	}

	boxRoundedStyle = func(r *lipgloss.Renderer) lipgloss.Style {
		b := lipgloss.RoundedBorder()
		return r.NewStyle().
			BorderStyle(b).
			Padding(0, 2)
	}

	promptStyle = func(r *lipgloss.Renderer) lipgloss.Style {
		return r.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ea76cb"))
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
		return boxRoundedStyle(r).Copy().BorderStyle(b)
	}
)

type modelFunc struct {
	textInput textinput.Model
	err       error
}

type modelInput struct {
	textInput textinput.Model
	err       error
}

func initURLInput(r *lipgloss.Renderer) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "r/gnoland/blog"
	ti.Focus()
	ti.CharLimit = 156
	ti.PromptStyle = promptStyle(r)
	ti.Prompt = gnoPrefix

	return ti
}

func initCommandInput(r *lipgloss.Renderer) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.CharLimit = 156
	ti.PromptStyle = promptStyle(r)
	ti.Prompt = "> "

	return ti
}

type model struct {
	kb     keys.Keybase
	render *lipgloss.Renderer
	banner string

	client *BroClient

	urlInput  textinput.Model
	listFuncs FuncListModel

	commandInput  textinput.Model
	commandFocus  bool
	zone          *zone.Manager
	ready         bool
	viewport      viewport.Model
	height, width int
	readonly      bool

	pageurls       map[string]string
	messageDisplay bool

	history *clist.List
	current *clist.Element
}

func (m model) Init() tea.Cmd {
	m.history.Init()
	return nil
}

// realm path surrounded by ansi escape sequences
var reUrlPattern = regexp.MustCompile(`(?mU)\x1b[^m]*m(?:(?:https?://)?gno.land)?(/[^\s]+)\x1b[^m]*m`)

func redirectWebPath(path string) string {
	if alias, ok := gnoweb.Aliases[path]; ok {
		return alias
	}

	if redirect, ok := gnoweb.Redirects[path]; ok {
		return redirect
	}

	return path
}

func (m model) findAndMarkURLs(body []byte) []byte {
	var buf bytes.Buffer
	lastIndex := 0

	indexes := reUrlPattern.FindAllSubmatchIndex(body, -1)
	for i, loc := range indexes {
		match := string(body[loc[0]:loc[1]])
		uri := string(body[loc[2]:loc[3]])
		markid := fmt.Sprintf("url_%d", i)

		// Write bytes before match
		buf.Write(body[lastIndex:loc[0]])

		// Write quoted URL
		buf.WriteString(m.zone.Mark(markid, match))
		m.pageurls[markid] = uri
		lastIndex = loc[1]
	}
	// Write remaining bytes
	buf.Write(body[lastIndex:])

	// Cleanup previous urls
	for i := len(indexes); i < len(m.pageurls); i++ {
		markid := fmt.Sprintf("url_%d", i)
		delete(m.pageurls, markid)
	}

	return buf.Bytes()
}

func (m model) fetchRenderView() (view []byte, err error) {
	path := m.urlInput.Value()
	rlmpath := gnoPrefix + path

	rlmpath, args, _ := strings.Cut(rlmpath, ":")
	res, err := m.client.Render(rlmpath, args)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch Render: %w", err)
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.viewport.Width),
	)

	if err != nil {
		return nil, fmt.Errorf("unable to get render view: %w", err)
	}

	view, err = r.RenderBytes(res)
	if err != nil {
		return nil, fmt.Errorf("uanble to render markdown view: %w", err)
	}

	return m.findAndMarkURLs(view), nil
}

func (m model) fetchFuncsList() (view []list.Item, err error) {
	path := m.urlInput.Value()
	rlmpath := gnoPrefix + path

	rlmpath, _, _ = strings.Cut(rlmpath, ":")
	funcs, err := m.client.Funcs(rlmpath)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch Render: %w", err)
	}

	items := make([]list.Item, 0, len(funcs))
	for _, fun := range funcs {
		if fun.FuncName != "Render" {
			items = append(items, itemFunc(fun))
		}
	}

	return items, nil
}

func (m model) getError(err error) string {
	f := wordwrap.NewWriter(m.viewport.Width)
	fmt.Fprintf(f, "error: %v", err)
	serr := f.String()
	return serr
}

func (m *model) moveToRealm(realm string) bool {
	// Trim prefix
	path := strings.TrimPrefix(realm, gnoPrefix)
	// redirect if any well known path
	path = redirectWebPath(path)
	// trim any slash
	path = strings.TrimPrefix(path, "/")
	if path == m.urlInput.Value() {
		return false
	}

	// Set uri input
	m.urlInput.SetValue(path)
	m.urlInput.CursorEnd()
	// Update render
	m.RenderUpdate()

	return true
}

func (m *model) updateHistory() {
	v := m.urlInput.Value()
	if m.history.Len() == 0 {
		m.current = m.history.PushBack(v)
		return
	}

	m.current = m.history.InsertAfter(v, m.current)
	if next := m.current.Next(); next != nil {
		m.history.Remove(next)
	}
}

func (m *model) moveHistoryForward() (string, bool) {
	if next := m.current.Next(); next != nil {
		m.current = next
		return m.current.Value.(string), true
	}
	return "", false
}

func (m *model) moveHistoryBackward() (string, bool) {
	if prev := m.current.Prev(); prev != nil {
		m.current = prev
		return m.current.Value.(string), true
	}
	return "", false
}

func (m *model) updateHistoryBackward() {
	v := m.urlInput.Value()
	if m.history.Len() == 0 {
		m.current = m.history.PushBack(v)
	} else {
		m.current = m.history.InsertAfter(v, m.current)
	}
}

func (m *model) RenderUpdate() {
	var err error
	render, err := m.fetchRenderView()
	if err != nil {
		f := wordwrap.NewWriter(m.viewport.Width)
		fmt.Fprintf(f, "error: %s", err)
		m.viewport.SetContent(f.String())
		return
	}

	m.viewport.SetContent(string(render))
	list, err := m.fetchFuncsList()
	if err != nil {
		f := wordwrap.NewWriter(m.viewport.Width)
		fmt.Fprintf(f, "error: %s", err)
		m.viewport.SetContent(f.String())
		return
	}

	if len(list) == 0 {
		return
	}

	m.listFuncs.Title = m.urlInput.Value()
	m.listFuncs.SetItems(list)
	m.listFuncs.SetSize(m.viewport.Width, 7)
}

func (m *model) ExtendCommandInput() bool {
	if m.commandInput.Focused() {
		if item, ok := m.listFuncs.SelectedItem().(itemFunc); ok {
			var value string
			if len(item.Params) > 0 {
				value = item.Title() + "("
			} else {
				value = item.Title() + "()"
			}

			currentValue := m.commandInput.Value()
			if len(value) > len(currentValue) && strings.HasPrefix(value, currentValue) {
				m.commandInput.SetValue(value)
				return true
			}
			m.commandInput.CursorEnd()

		}
	}

	return false
}

type UpdateRenderMsg struct {
	realmPath string
}

// XXX: it's bit messy here, need some rework
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	if m.banner != "" {
		if _, ok := msg.(tea.KeyMsg); ok {
			m.banner = ""
			return m, nil
		}
	}

	switch msg := msg.(type) {
	case UpdateRenderMsg:
		if msg.realmPath != "" {
			m.moveToRealm(msg.realmPath)
		} else {
			m.RenderUpdate()
		}

	case tea.MouseMsg:
		if msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown ||
			msg.Button == tea.MouseButtonWheelLeft || msg.Button == tea.MouseButtonWheelRight {
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd // stop here to avoid update input view
		}

		if msg.Action == tea.MouseActionRelease {
			switch {
			case m.zone.Get("prev_button").InBounds(msg):
				if path, ok := m.moveHistoryBackward(); ok {
					m.moveToRealm(path)
				}
			case m.zone.Get("next_button").InBounds(msg):
				if path, ok := m.moveHistoryForward(); ok {
					m.moveToRealm(path)
				}
			case m.zone.Get("home_button").InBounds(msg):
				if m.moveToRealm("gno.land/r/gnoland/home") {
					m.updateHistory()
				}

			case m.zone.Get("url_input").InBounds(msg):
				m.commandInput.Blur()
				cmds = append(cmds, m.urlInput.Focus())
				m.commandFocus = false
			case !m.readonly && m.zone.Get("command_input").InBounds(msg):
				m.urlInput.Blur()
				cmds = append(cmds, m.commandInput.Focus())
				m.commandFocus = true
			default:
				for mark := range m.pageurls {
					if !m.zone.Get(mark).InBounds(msg) {
						continue
					}

					if uri := m.pageurls[mark]; uri != "" {
						if m.moveToRealm(uri) {
							m.updateHistory()
						}
					}

					break
				}
			}
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "alt+down":
			if m.urlInput.Focused() && !m.readonly {
				m.urlInput.Blur()
				cmds = append(cmds, m.commandInput.Focus())
				m.commandFocus = true
			}
		case "alt+up":
			if m.commandInput.Focused() {
				m.commandInput.Blur()
				cmds = append(cmds, m.urlInput.Focus())
				m.commandFocus = false
			}
		case "tab":
			m.ExtendCommandInput()
		case "down":
			if m.commandFocus {
				m.listFuncs.CursorDown()
			}
		case "up":
			if m.commandFocus {
				m.listFuncs.CursorUp()
			}
		case "alt+r":
			m.RenderUpdate()
		case "enter":
			if m.commandInput.Focused() && !m.messageDisplay {
				if len(m.listFuncs.Items()) > 1 {
					m.ExtendCommandInput()
					break
				}

				path := m.urlInput.Value()
				rlmpath := gnoPrefix + path
				res, err := m.client.Call(rlmpath, m.commandInput.Value())
				if err != nil {
					content := fmt.Sprintf("%s\n\npress [enter] to dismiss", m.getError(err))
					m.viewport.SetContent(content)
					m.messageDisplay = true
				} else {
					if strings.TrimSpace(string(res)) == "" {
						m.RenderUpdate()
						m.messageDisplay = false
					} else {
						m.viewport.SetContent(fmt.Sprintf("%s\n\npress [enter] to dismiss", string(res)))
						m.messageDisplay = true
					}
				}

				m.listFuncs.Erase()
				break
			}

			if m.messageDisplay || m.urlInput.Focused() {
				m.listFuncs.Erase()
				m.RenderUpdate()
				if m.current.Value.(string) != m.urlInput.Value() {
					m.updateHistory()
				}

				m.messageDisplay = false
			}

		case "ctrl+c", "esc":
			return m, tea.Quit
		default:
			if m.urlInput.Focused() {
				m.urlInput, cmd = m.urlInput.Update(msg)
			} else {
				m.commandInput, cmd = m.commandInput.Update(msg)
			}
			// handle url input

		}

	case tea.WindowSizeMsg:
		m.width = msg.Width

		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.MouseWheelEnabled = true
			m.viewport.MouseWheelDelta = 1
			m.viewport.HighPerformanceRendering = useHighPerformanceRenderer
			m.ready = true

			// This is only necessary for high performance rendering, which in
			// most cases you won't need.
			//
			// Render the viewport one line below the header.
			m.viewport.YPosition = headerHeight + 1

			if value := m.urlInput.Value(); value != "" {
				m.RenderUpdate()
				m.updateHistory()
			}
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

		m.height = m.viewport.Height
		if !m.urlInput.Focused() && len(m.listFuncs.Items()) > 0 {
			m.viewport.Height = m.height - lipgloss.Height(m.listFuncsView())
		}

		if useHighPerformanceRenderer {
			// Render (or re-render) the whole viewport. Necessary both to
			// initialize the viewport and when the window is resized.
			//
			// This is needed for high-performance rendering only.
			cmds = append(cmds, viewport.Sync(m.viewport))
		}
	}

	// m.listFuncs, cmd = m.listFuncs.Update(msg)
	// cmds = append(cmds, cmd)

	if v := m.commandInput.Value(); v != "" {
		m.listFuncs.FilterItems(v)
	} else {
		m.listFuncs.Reset()
	}

	if m.commandFocus && len(m.listFuncs.Items()) > 0 {
		m.viewport.Height = m.height - lipgloss.Height(m.listFuncsView())
	} else {
		m.viewport.Height = m.height
	}

	// Handle keyboard and mouse events in the viewport
	if m.urlInput.Focused() {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "+"
	}

	if m.banner != "" {
		banner := m.render.NewStyle().Padding(1, 5).
			Border(lipgloss.DoubleBorder(), true, false, true).
			Render(m.banner)

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, banner)
	}

	mainView := fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.bodyView(), m.footerView())
	return m.zone.Scan(mainView)
}

func (m model) listFuncsView() string {
	return boxRoundedStyle(m.render).
		Render(m.listFuncs.View())
}

func (m model) bodyView() string {
	return m.viewport.View()
}

func (m model) headerView() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.navView(), m.urlView())
}

func (m model) urlView() string {
	return m.zone.Mark("url_input", boxRoundedStyle(m.render).
		Width(m.viewport.Width-2).
		Render(m.urlInput.View()))
}

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

	spaceWidth := m.width / 3
	return lipgloss.JoinHorizontal(lipgloss.Left,
		m.render.NewStyle().Width(spaceWidth).Padding(0, 1).
			Render(lipgloss.JoinHorizontal(lipgloss.Left,
				m.zone.Mark("prev_button", prev),
				m.zone.Mark("next_button", next),
			)),
		m.render.PlaceHorizontal(spaceWidth, lipgloss.Center, "Gno.Land"),
		m.render.PlaceHorizontal(spaceWidth, lipgloss.Right,
			m.zone.Mark("home_button", home)),
	)

}

func (m model) footerView() string {
	if m.readonly {
		return ""
	}

	info := infoStyle(m.render).Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
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
