package components

import (
	"fmt"
	"io"
	"time"
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

func FormatRelativeTimeSince(t time.Time) string {
	diff := time.Since(t)

	units := []struct {
		unit  time.Duration
		label string
	}{
		{time.Minute, "minute"},
		{time.Hour, "hour"},
		{24 * time.Hour, "day"},
		{7 * 24 * time.Hour, "week"},
		{30 * 24 * time.Hour, "month"},
		{365 * 24 * time.Hour, "year"},
	}

	for i := len(units) - 1; i >= 0; i-- {
		u := units[i]
		if diff >= u.unit {
			value := int(diff / u.unit)
			if value == 1 {
				return fmt.Sprintf("1 %s ago", u.label)
			}
			return fmt.Sprintf("%d %ss ago", value, u.label)
		}
	}

	return "just now"
}
