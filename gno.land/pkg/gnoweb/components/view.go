package components

import (
	"fmt"
	"io"
)

// ViewType represents the type of a view component.
type ViewType string

// View represents a UI component with a type and underlying component.
type View struct {
	Type ViewType
	// SkipTargetInBody marks views that render their rail before the content
	// and so carry #main-content themselves, past the rail; <main> carries it
	// otherwise. Set by the constructor that picks the template.
	SkipTargetInBody bool
	Component
}

func (v *View) String() string {
	return string(v.Type)
}

// Render renders the view to the provided writer.
func (v *View) Render(w io.Writer) error {
	if err := v.Component.Render(w); err != nil {
		return fmt.Errorf("view %q error: %w", string(v.Type), err)
	}

	return nil
}

// NewTemplateView creates a new View with a template component and data.
func NewTemplateView(typ ViewType, name string, data any) *View {
	return &View{
		Type:      typ,
		Component: NewTemplateComponent(name, data),
	}
}
