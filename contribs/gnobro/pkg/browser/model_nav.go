package browser

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

func (m *model) moveToRealm(realm string) tea.Cmd {
	path := cleanupRealmPath(m.urlPrefix, realm)

	// Set uri input
	m.urlInput.SetValue(path)
	m.urlInput.CursorEnd()

	// return command update
	return tea.Sequence(send(RefreshRealmMsg()), m.urlInput.Focus())
}

func (m *model) updateHistory() {
	v := m.urlInput.Value()
	if m.history.Len() == 0 {
		m.current = m.history.PushBack(v)
		return
	}

	m.current = m.history.InsertAfter(v, m.current)
	for next := m.current.Next(); next != nil; {
		m.history.Remove(next)
		next = m.current.Next()
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

func (m model) fetchRenderView(path string) (view []byte, err error) {
	rlmpath, args, _ := strings.Cut(path, ":")
	res, err := m.client.Render(rlmpath, args)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch Render: %w", err)
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(CatppuccinStyleConfig), // XXX: use gno custom theme
		glamour.WithWordWrap(m.viewport.Width),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get render view: %w", err)
	}

	view, err = r.RenderBytes(res)
	if err != nil {
		return nil, fmt.Errorf("uanble to render markdown view: %w", err)
	}

	return view, nil
}
