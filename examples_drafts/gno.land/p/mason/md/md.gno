package md

import (
	"strings"
)

type MD struct {
	elements []string
}

func New() *MD {
	return &MD{elements: []string{}}
}

func (m *MD) H1(text string) {
	m.elements = append(m.elements, "# "+text)
}

func (m *MD) H3(text string) {
	m.elements = append(m.elements, "### "+text)
}

func (m *MD) P(text string) {
	m.elements = append(m.elements, text)
}

func (m *MD) Code(text string) {
	m.elements = append(m.elements, "  ```\n"+text+"\n```\n")
}

func (m *MD) Im(path string, caption string) {
	m.elements = append(m.elements, "!["+caption+"]("+path+" \""+caption+"\")")
}

func (m *MD) Bullet(point string) {
	m.elements = append(m.elements, "- "+point)
}

func Link(text, url string, title ...string) string {
	if len(title) > 0 && title[0] != "" {
		return "[" + text + "](" + url + " \"" + title[0] + "\")"
	}
	return "[" + text + "](" + url + ")"
}

func (m *MD) Render() string {
	return strings.Join(m.elements, "\n\n")
}
