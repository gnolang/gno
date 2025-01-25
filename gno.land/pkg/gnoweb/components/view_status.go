package components

const StatusViewType ViewType = "status-view"

// StatusData holds the dynamic fields for the "status" template
type StatusData struct {
	Title      string
	Body       string
	ButtonURL  string
	ButtonText string
}

// StatusErrorComponent returns a view for error scenarios
func StatusErrorComponent(message string) *View {
	return NewTemplateView(
		StatusViewType,
		"status",
		StatusData{
			Title:      "Error",
			Body:       message,
			ButtonURL:  "/",
			ButtonText: "Go Back Home",
		},
	)
}

// StatusNoRenderComponent returns a view for non-error notifications
func StatusNoRenderComponent(pkgPath string) *View {
	return NewTemplateView(
		StatusViewType,
		"status",
		StatusData{
			Title:      "No Render",
			Body:       "The developer of this realm did not create a Render() function.",
			ButtonURL:  pkgPath + "$source",
			ButtonText: "View Realm Source",
		},
	)
}
