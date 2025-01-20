package components

import "io"

type RedirectData struct {
	To            string
	WithAnalytics bool
}

func RenderRedirectComponent(w io.Writer, data RedirectData) error {
	return tmpl.ExecuteTemplate(w, "renderRedirect", data)
}
