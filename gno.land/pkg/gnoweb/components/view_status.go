package components

const StatusViewType ViewType = "status-view"

type StatusData struct {
	Message string
}

func RenderStatusComponent(message string) *View {
	return NewTemplateView(StatusViewType, "status", StatusData{message})
}
