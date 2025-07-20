package browser

import (
	"bytes"
	clist "container/list"
	"errors"
	"fmt"
	"log/slog"
	gopath "path"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/muesli/reflow/wordwrap"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/log"
)

var promptStyle = func(r *lipgloss.Renderer) lipgloss.Style {
	return r.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#dd7878"))
}

var ErrEmptyRenderer = errors.New("empty rendrer")

type Config struct {
	URLPrefix       string
	URLDefaultValue string
	Logger          *slog.Logger
	Renderer        *lipgloss.Renderer
	Readonly        bool
	Banner          ModelBanner
	DevMode         bool
}

const DefaultGnoLandPrefix = "gno.land/"

func DefaultConfig() Config {
	return Config{
		Logger:          log.NewNoopLogger(),
		URLPrefix:       DefaultGnoLandPrefix,
		Renderer:        lipgloss.DefaultRenderer(),
		URLDefaultValue: "gnoland/home",
	}
}

type model struct {
	render *lipgloss.Renderer
	client *NodeClient
	logger *slog.Logger

	// misc
	banner          ModelBanner
	bannerDiscarded bool

	// Viewport
	zone           *zone.Manager
	ready          bool
	viewport       viewport.Model
	height, width  int
	readonly       bool
	messageDisplay bool

	// Nav
	taskLoader LoaderModel

	pageurls map[string]string
	history  *clist.List
	current  *clist.Element

	// Url
	urlInput  textinput.Model
	urlPrefix string

	// Commands
	listFuncs    FuncListModel
	commandInput textinput.Model
	commandFocus bool

	// Dev
	devMode         bool
	devClientStatus ClientStatus
}

func initURLInput(prefix string, r *lipgloss.Renderer) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "r/gnoland/blog" // XXX: Use as example, customize this ?
	ti.Focus()
	ti.CharLimit = 156
	ti.PromptStyle = promptStyle(r)
	ti.Prompt = prefix + "/"

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

func New(cfg Config, client *gnoclient.Client) tea.Model {
	renderer := lipgloss.DefaultRenderer()
	if cfg.Renderer != nil {
		renderer = cfg.Renderer
	}

	// Setup url input
	urlinput := initURLInput(cfg.URLPrefix, renderer)

	path := cleanupRealmPath(cfg.URLPrefix, cfg.URLDefaultValue)
	urlinput.SetValue(path)

	// Setup cmd input
	cmdinput := initCommandInput(renderer)

	// XXX: Customize this
	base := gnoclient.BaseTxCfg{
		GasFee:    "1000000ugnot",
		GasWanted: 2000000,
	}

	nodeclient := NewNodeClient(cfg.Logger, base, client)
	return &model{
		logger:     cfg.Logger,
		render:     cfg.Renderer,
		readonly:   cfg.Readonly,
		client:     nodeclient,
		taskLoader: newLoaderModel(),

		banner:          cfg.Banner,
		bannerDiscarded: cfg.Banner.Empty(),

		urlInput:  urlinput,
		urlPrefix: cfg.URLPrefix,

		commandInput: cmdinput,
		listFuncs:    newFuncList(),

		zone:     zone.New(),
		pageurls: map[string]string{},
		history:  clist.New(),

		devMode: cfg.DevMode,
	}
}

func (m model) Init() tea.Cmd {
	m.history.Init()
	return m.banner.Init()
}

type fetchRealmMsg struct {
	realmPath string
}

func FetchRealmMsg(path string) tea.Msg {
	return fetchRealmMsg{path}
}

func RefreshRealmMsg() tea.Msg {
	return fetchRealmMsg{""}
}

type clientStatusUpdateMsg struct {
	status ClientStatus
	remote string
}

func DevClientStatusUpdateMsg(s ClientStatus, remote string) tea.Msg {
	return clientStatusUpdateMsg{s, remote}
}

type renderUpdateMsg struct {
	Render []byte
	Funcs  vm.FunctionSignatures
	Error  error
}

func (m *model) RenderUpdate(path string) tea.Cmd {
	return func() tea.Msg {
		var msg renderUpdateMsg
		var err error
		msg.Render, err = m.fetchRenderView(path)
		if err != nil {
			msg.Error = fmt.Errorf("unable to fetch view: %w", err)
			return msg
		}

		msg.Funcs, err = m.fetchFuncsList(path)
		if err != nil {
			msg.Error = fmt.Errorf("unable to fetch function list: %w", err)
			return msg
		}

		return msg
	}
}

type execCommandRequestMsg struct {
	Path    string
	Command string
}

func (m *model) ExecCommandRequest(path, command string) tea.Cmd {
	return func() tea.Msg {
		return execCommandRequestMsg{path, command}
	}
}

type execCommandMsg struct {
	Response []byte
	Error    error
}

func (m *model) ExecCommand(path, command string) tea.Cmd {
	return func() tea.Msg {
		res, err := m.client.Call(path, command)
		return execCommandMsg{res, err}
	}
}

func (m *model) ExtendCommandInput() bool {
	if !m.commandInput.Focused() {
		return false
	}

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

		// Put cursor at the end
		m.commandInput.CursorEnd()
	}

	return false
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case fetchRealmMsg:
		if msg.realmPath != "" {
			return m, tea.Sequence(m.taskLoader.Add(1), m.moveToRealm(msg.realmPath))
		}

		// If no realm path is given simply refresh the current realm
		path := m.getCurrentPath()
		m.logger.Info("rendering realm", "path", path)

		return m, tea.Sequence(m.taskLoader.Add(1), m.RenderUpdate(path))

	case execCommandRequestMsg:
		m.logger.Info("requesting command", "path", msg.Path, "cmd", msg.Command)
		cmd = m.ExecCommand(msg.Path, msg.Command)
		return m, tea.Sequence(m.taskLoader.Add(1), cmd)

	case execCommandMsg:
		m.taskLoader.Done()

		// If any error, display it as message.
		if msg.Error != nil {
			m.logger.Warn("command exec", "error", msg.Error)

			content := wordwrap.NewWriter(m.viewport.Width)
			fmt.Fprint(content, msg.Error.Error())
			fmt.Fprintf(content, "\n\npress [enter] to dismiss error\n")
			m.viewport.SetContent(content.String())
			m.messageDisplay = true
			return m, nil
		}

		// If any response, display it as message.
		if res := bytes.TrimSpace(msg.Response); len(res) > 0 {
			m.logger.Info("command exec", "res", string(res))

			content := wordwrap.NewWriter(m.viewport.Width)
			content.Write(res)
			fmt.Fprintf(content, "\n\npress [enter] to dismiss message\n")
			m.viewport.SetContent(content.String())
			m.messageDisplay = true
			return m, nil
		}

		// If no error or empty response is returned, simply refresh the page.
		m.messageDisplay = false
		return m, send(RefreshRealmMsg())

	case renderUpdateMsg:
		m.taskLoader.Done()

		var content string
		if err := msg.Error; err != nil {
			m.logger.Warn("render", "error", msg.Error)
			// Write error to the frame
			content = fmt.Sprintf("ERROR: %s", err.Error())
		} else {
			content = string(m.findAndMarkURLs(msg.Render))
		}

		if len(msg.Funcs) > 0 {
			items := make([]list.Item, 0, len(msg.Funcs))
			for _, fun := range msg.Funcs {
				if fun.FuncName != "Render" {
					items = append(items, itemFunc(fun))
				}
			}
			m.listFuncs.SetItems(items)
			m.listFuncs.FilterItems(m.commandInput.Value())

			// Update funcs list
			m.listFuncs.Title = m.urlInput.Value()
			m.listFuncs.SetSize(m.viewport.Width, 7)
		}

		m.viewport.SetContent(content)
		return m, cmd

	case SpinnerTickMsg:
		if m.taskLoader.Active() {
			m.taskLoader, cmd = m.taskLoader.Update(msg)
		}

	case clientStatusUpdateMsg:
		m.devClientStatus = msg.status
		return m, cmd

	case tea.MouseMsg:
		cmd = m.updateMouse(msg)

		// Fallback on viewport
		if cmd == nil {
			m.viewport, cmd = m.viewport.Update(msg)
		}

		return m, cmd

	case tea.KeyMsg:
		cmd = m.updateKey(msg)
		m.listFuncs.FilterItems(m.commandInput.Value())
		if !m.readonly && cmd == nil {
			m.listFuncs, cmd = m.listFuncs.Update(msg)
		}

		// Fallback on list funcs update
		if cmd == nil {
			m.viewport, cmd = m.viewport.Update(msg)
		}

		// Fallback on viewport update
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width

		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.MouseWheelEnabled = true
			m.viewport.MouseWheelDelta = 1
			m.ready = true
			m.viewport.YPosition = headerHeight + 1

			if value := m.urlInput.Value(); value != "" {
				cmd = send(RefreshRealmMsg())
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

		return m, cmd
	}

	// Update other models
	cmds := []tea.Cmd{cmd}

	if !m.bannerDiscarded {
		var bannerCmd tea.Cmd
		m.banner, bannerCmd = m.banner.Update(msg)
		cmds = append(cmds, bannerCmd)
	}

	var viewCmd tea.Cmd
	m.viewport, viewCmd = m.viewport.Update(msg)
	cmds = append(cmds, viewCmd)

	var funcCmd tea.Cmd
	m.listFuncs, funcCmd = m.listFuncs.Update(msg)
	cmds = append(cmds, funcCmd)

	return m, tea.Batch(cmds...)
}

func (m *model) updateKey(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	if !m.bannerDiscarded {
		switch key := msg.String(); key {
		case "ctrl+c":
			return tea.Quit
		case "enter":
			m.bannerDiscarded = true
		}
		// Discard other input while banner is active
		return nil
	}

	switch msg.String() {
	case "alt+down":
		if m.urlInput.Focused() && !m.readonly {
			m.urlInput.Blur()
			cmd = m.commandInput.Focus()
			m.commandFocus = true
		}
	case "alt+up":
		if m.commandInput.Focused() {
			m.commandInput.Blur()
			cmd = m.urlInput.Focus()
			m.commandFocus = false
		}
	case "tab":
		if m.commandInput.Focused() {
			m.ExtendCommandInput()
		}
	case "alt+r":
		cmd = send(RefreshRealmMsg())
	case "enter":
		// Update command on focus
		if m.commandInput.Focused() && !m.messageDisplay {
			if len(m.listFuncs.Items()) == 1 {
				path := m.getCurrentPath()
				cmd = m.ExecCommand(path, m.commandInput.Value())
			} else {
				m.ExtendCommandInput()
			}

			break
		}

		// Update url on focus
		if m.messageDisplay || m.urlInput.Focused() {
			m.listFuncs.Erase()

			cmd = m.moveToRealm(m.urlInput.Value())
			if m.current.Value.(string) != m.urlInput.Value() {
				m.updateHistory()
			}

			// Discard message
			m.messageDisplay = false
		}

	case "ctrl+c", "esc":
		return tea.Quit
	default:
		// handle url input
		if m.urlInput.Focused() {
			m.urlInput, cmd = m.urlInput.Update(msg)
		}

		if m.commandInput.Focused() {
			// handle command input
			m.commandInput, cmd = m.commandInput.Update(msg)
		}
	}

	return cmd
}

func (m *model) updateMouse(msg tea.MouseMsg) tea.Cmd {
	if msg.Action != tea.MouseActionRelease {
		return nil
	}

	var cmd tea.Cmd

	switch {
	case m.zone.Get("prev_button").InBounds(msg):
		if path, ok := m.moveHistoryBackward(); ok {
			cmd = m.moveToRealm(path)
		}
	case m.zone.Get("next_button").InBounds(msg):
		if path, ok := m.moveHistoryForward(); ok {
			cmd = m.moveToRealm(path)
		}
	case m.zone.Get("home_button").InBounds(msg):
		if cmd = m.moveToRealm("gno.land/r/gnoland/home"); cmd != nil {
			m.updateHistory()
		}

	case m.zone.Get("url_input").InBounds(msg):
		m.commandInput.Blur()
		cmd = m.urlInput.Focus()
		m.commandFocus = false
	case !m.readonly && m.zone.Get("command_input").InBounds(msg):
		m.urlInput.Blur()
		cmd = m.commandInput.Focus()
		m.commandFocus = true
	default:
		for mark := range m.pageurls {
			if !m.zone.Get(mark).InBounds(msg) {
				continue
			}

			if uri := m.pageurls[mark]; uri != "" {
				if cmd = m.moveToRealm(uri); cmd != nil {
					m.updateHistory()
					break
				}
			}
		}
	}

	return cmd
}

// realm path surrounded by ansi escape sequences
var reUrlPattern = regexp.MustCompile(`(?mU)\x1b[^m]*m(?:(?:https?://)?gno.land)?(/[^\s]+)\x1b[^m]*m`)

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

func (m model) fetchFuncsList(path string) (view vm.FunctionSignatures, err error) {
	rlmpath, _, _ := strings.Cut(path, ":")
	funcs, err := m.client.Funcs(rlmpath)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch Render: %w", err)
	}

	return funcs, nil
}

func (m *model) getCurrentPath() string {
	path := strings.Trim(m.urlInput.Value(), "/")
	if len(path) == 0 {
		return m.urlPrefix
	}

	return gopath.Join(m.urlPrefix, path)
}

func send(msg tea.Msg) tea.Cmd {
	return func() tea.Msg { return msg }
}
