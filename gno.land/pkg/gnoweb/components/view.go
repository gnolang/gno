package components

import (
	"fmt"
	"io"
)

type ViewType string

type View struct {
	Type ViewType
	Component
}

func (v *View) String() string {
	return string(v.Type)
}

func (v *View) Render(w io.Writer) error {
	if err := v.Component.Render(w); err != nil {
		return fmt.Errorf("view %q error: %w", string(v.Type), err)
	}

	return nil
}

func NewTemplateView(typ ViewType, name string, data any) *View {
	return &View{
		Type:      typ,
		Component: NewTemplateComponent(name, data),
	}
}
