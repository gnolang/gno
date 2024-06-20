package main

import (
	"bytes"
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
	boxRoundedStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		return lipgloss.NewStyle().
			BorderStyle(b).
			Padding(0, 2)
	}()

	inputStyleLeft = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().
			BorderStyle(b).
			Padding(0, 2)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return boxRoundedStyle.Copy().BorderStyle(b)
	}()
)

// func (i itemFunc) Description() string {
// 	var str strings
// 	return i.Params
// }

type modelFunc struct {
	textInput textinput.Model
	err       error
}

type modelInput struct {
	textInput textinput.Model
	err       error
}

func initURLInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "demo/foo20"
	ti.Focus()
	ti.CharLimit = 156
	ti.PromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF06B7"))
	ti.Prompt = gnoPrefix

	return ti
}

func initCommandInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.CharLimit = 156
	ti.PromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF06B7"))
	ti.Prompt = "> "

	return ti
}

type model struct {
	name, pass string
	kb         keys.Keybase

	client *BroClient

	urlInput  textinput.Model
	listFuncs FuncListModel

	commandInput textinput.Model
	commandFocus bool
	zone         *zone.Manager
	ready        bool
	viewport     viewport.Model
	height       int

	pageurls       map[string]string
	messageDisplay bool
}

func (m model) Init() tea.Cmd {
	return nil
}

var urlPattern = regexp.MustCompile(`(?m)/(?:p|r)/[^[;\s]+`)

func (m model) FindAndMarkURLs(body []byte) []byte {
	var buf bytes.Buffer
	lastIndex := 0

	indexes := urlPattern.FindAllIndex(body, -1)
	for i, loc := range indexes {
		markid := fmt.Sprintf("url_%d", i)

		// Write bytes before match
		buf.Write(body[lastIndex:loc[0]])

		// Write quoted URL
		u := string(body[loc[0]:loc[1]])
		buf.WriteString(m.zone.Mark(markid, u))
		m.pageurls[markid] = u
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

	return m.FindAndMarkURLs(view), nil
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

		}
	}

	return false
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown ||
			msg.Button == tea.MouseButtonWheelLeft || msg.Button == tea.MouseButtonWheelRight {
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd // stop here to avoid update input view
		}

		if m.zone.Get("url_input").InBounds(msg) {
			m.commandInput.Blur()
			cmds = append(cmds, m.urlInput.Focus())
			m.commandFocus = false
		} else if m.zone.Get("command_input").InBounds(msg) {
			m.urlInput.Blur()
			cmds = append(cmds, m.commandInput.Focus())
			m.commandFocus = true
		} else {
			for mark := range m.pageurls {
				if !m.zone.Get(mark).InBounds(msg) {
					continue
				}

				if uri := m.pageurls[mark]; uri != "" {
					uri = strings.TrimPrefix(uri, "/")
					m.urlInput.SetValue(uri)
					m.urlInput.CursorEnd()
					m.RenderUpdate()
				}

				break
			}
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "alt+down", "alt+up", "tab":
			if m.urlInput.Focused() {
				m.urlInput.Blur()
				cmds = append(cmds, m.commandInput.Focus())
				m.commandFocus = true
			} else if m.commandInput.Focused() {
				if msg.String() == "tab" && m.ExtendCommandInput() {
					break
				} else {
					m.commandInput.Blur()
					cmds = append(cmds, m.urlInput.Focus())
					m.commandFocus = false
				}
			}
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
		return "\n  Initializing..."
	}

	mainView := fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.bodyView(), m.footerView())
	return m.zone.Scan(mainView)
}

func (m model) listFuncsView() string {
	return boxRoundedStyle.
		Render(m.listFuncs.View())
}

func (m model) bodyView() string {
	return m.viewport.View()
}

func (m model) headerView() string {
	return m.zone.Mark("url_input", boxRoundedStyle.
		Width(m.viewport.Width-2).
		Render(m.urlInput.View()))
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	command := m.zone.Mark("command_input", inputStyleLeft.
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
