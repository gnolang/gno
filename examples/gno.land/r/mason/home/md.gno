package home

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

func (m *MD) Im(path string, caption string) {
	m.elements = append(m.elements, caption)
}

func (m *MD) Render() string {
	return strings.Join(m.elements, "\n\n")
}