package components

const RedirectViewType = "redirect-view"

type RedirectData struct {
	To            string
	WithAnalytics bool
}

func RenderRedirectView(data RedirectData) *View {
	return NewTemplateView(RedirectViewType, "renderRedirect", data)
}
