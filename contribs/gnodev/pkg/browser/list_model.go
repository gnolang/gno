package browser

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/sahilm/fuzzy"
)

var (
	listItemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	listSelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	listPaginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
)

type FuncListModel struct {
	list.Model
	items []list.Item
}

func (m *FuncListModel) SetItems(items []list.Item) {
	m.Model.SetItems(items)
	m.items = items
}

func (m FuncListModel) Update(msg tea.Msg) (FuncListModel, tea.Cmd) {
	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	return m, cmd
}

func (m *FuncListModel) OriginItems() []list.Item {
	return m.items
}

func (m *FuncListModel) Erase() {
	m.Model.SetItems([]list.Item{})
}

func (m *FuncListModel) Reset() {
	m.Model.SetItems(m.items)
}

func (m *FuncListModel) FilterItems(pattern string) {
	if pattern == "" {
		m.Reset()
		return
	}

	i := strings.IndexRune(pattern, '(')
	if i > 0 {
		pattern = pattern[:i]
	}

	data := make([]string, len(m.items))
	for i, item := range m.items {
		data[i] = item.FilterValue()
	}

	ranks := fuzzy.Find(pattern, data)
	sort.Stable(ranks)
	if len(ranks) > 0 && i > 0 {
		m.Model.SetItems([]list.Item{m.items[ranks[0].Index]})
		return
	}

	items := make([]list.Item, len(ranks))
	for i, r := range ranks {
		items[i] = m.items[r.Index]
	}

	m.Model.SetItems(items)
}

type itemFunc vm.FunctionSignature

func (i itemFunc) Name() string        { return i.FuncName }
func (i itemFunc) Title() string       { return i.Name() }
func (i itemFunc) Description() string { return i.FuncName }
func (i itemFunc) FilterValue() string { return i.FuncName }

type itemFuncsDelegate struct{}

func (d itemFuncsDelegate) Height() int                             { return 1 }
func (d itemFuncsDelegate) Spacing() int                            { return 0 }
func (d itemFuncsDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemFuncsDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	fun, ok := listItem.(itemFunc)
	if !ok {
		return
	}

	maxw := m.Width() - 10

	var proto strings.Builder
	fmt.Fprintf(&proto, "%s(", fun.FuncName)
	for j, param := range fun.Params {
		if j != 0 {
			fmt.Fprint(&proto, ", ")
		}

		fmt.Fprintf(&proto, "%s %s", param.Name, param.Type)
	}
	fmt.Fprint(&proto, ")")

	switch len(fun.Results) {
	case 0: // none
	case 1:
		fmt.Fprintf(&proto, " %s", fun.Results[0].Type)
	default:
		fmt.Fprint(&proto, " (")
		for j, res := range fun.Results {
			if j != 0 {
				fmt.Fprint(&proto, ", ")
			}

			fmt.Fprint(&proto, res.Type)
		}
		fmt.Fprint(&proto, ")")
	}

	fn := listItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return listSelectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	str := proto.String()
	if len(str) > maxw {
		str = str[:maxw-3] + "..."
	}

	fmt.Fprint(w, fn(str))
}

func newFuncList() FuncListModel {
	l := list.New([]list.Item{}, &itemFuncsDelegate{}, 0, 0)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.Styles.PaginationStyle = listPaginationStyle
	return FuncListModel{
		Model: l,
		items: l.Items(),
	}
}
