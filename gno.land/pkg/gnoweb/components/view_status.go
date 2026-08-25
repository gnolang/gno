package components

const StatusViewType ViewType = "status-view"

// StatusData holds the dynamic fields for the "status" template
type StatusData struct {
	Title      string
	Body       string
	ButtonURL  string
	ButtonText string
}

// StatusErrorComponent returns a view for error scenarios with a message.
func StatusErrorComponent(message string) *View {
	return NewTemplateView(
		StatusViewType,
		"status",
		StatusData{
			Title:      "Error: " + message,
			Body:       "Something went wrong.",
			ButtonURL:  "/",
			ButtonText: "Go Back Home",
		},
	)
}

// StatusNoRenderComponent returns a view for non-error notifications when Render() is not implemented.
// StatusPendingApprovalComponent is shown for a package that was submitted but
// has not been approved yet.
//
// Distinct from the not-found status on purpose. Under the "inert" code
// submission policy those two look identical from every other query -- the
// package is stored, but nothing that reads the live key space can see it --
// and telling a creator who has already paid to submit that their path does not
// exist is the confusing half of that.
func StatusPendingApprovalComponent(pkgPath, reason string) *View {
	body := "This package has been submitted. Its source is not readable and it cannot be called until it is enabled."
	if reason != "" {
		body += " Reason: " + reason + "."
	}
	return NewTemplateView(
		StatusViewType,
		"status",
		StatusData{
			Title:      "Not Yet Enabled",
			Body:       body,
			ButtonURL:  "/",
			ButtonText: "Go Back Home",
		},
	)
}

func StatusNoRenderComponent(pkgPath string) *View {
	return NewTemplateView(
		StatusViewType,
		"status",
		StatusData{
			Title:      "No Render",
			Body:       "This realm does not implement a Render() function.",
			ButtonURL:  pkgPath + "$source",
			ButtonText: "View Realm Source",
		},
	)
}
