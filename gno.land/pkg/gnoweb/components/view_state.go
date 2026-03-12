package components

const StateViewType ViewType = "state-view"

// StateData holds data for rendering the state explorer view.
type StateData struct {
	PkgPath   string
	NodesJSON string // JSON string embedded in the template for initial render.
}

type stateViewParams struct {
	PkgPath   string
	NodesJSON string
}

// StateView creates a new View for the state explorer.
func StateView(data StateData) *View {
	return NewTemplateView(StateViewType, "renderState", stateViewParams{
		PkgPath:   data.PkgPath,
		NodesJSON: data.NodesJSON,
	})
}
