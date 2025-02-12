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

func (m *MD) h1(text string) {
	m.elements = append(m.elements, "# "+text)
}

func (m *MD) h3(text string) {
	m.elements = append(m.elements, "### "+text)
}

func (m *MD) p(text string) {
	m.elements = append(m.elements, text)
}

func (m *MD) code(text string) {
	m.elements = append(m.elements, "  ```\n"+text+"\n```\n")
}

func (m *MD) im(path string, caption string) {
	m.elements = append(m.elements, "!["+caption+"]("+path+" \""+caption+"\")")
}

func (m *MD) bullet(point string) {
	m.elements = append(m.elements, "- "+point)
}

func link(text, url string, title ...string) string {
    if len(title) > 0 && title[0] != "" {
        return "["+text+"]("+url+" \""+title[0]+"\")"
    }
    return "["+text+"]("+url+")"
}

func (m *MD) Render() string {
	return strings.Join(m.elements, "\n\n")
}
