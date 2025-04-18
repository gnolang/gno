package components

import (
	"io"
)

type Component interface {
	Render(w io.Writer) error
}

type TemplateComponent struct {
	name string
	data any
}

func (c *TemplateComponent) Render(w io.Writer) error {
	return tmpl.ExecuteTemplate(w, c.name, c.data)
}

func NewTemplateComponent(name string, data any) Component {
	return &TemplateComponent{name: name, data: data}
}

type readerComponent struct {
	io.Reader
}

func NewReaderComponent(reader io.Reader) Component {
	return &readerComponent{reader}
}

func (c *readerComponent) Render(w io.Writer) (err error) {
	_, err = io.Copy(w, c)
	return err
}
