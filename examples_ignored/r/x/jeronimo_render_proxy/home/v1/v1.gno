package v1

import "gno.land/r/x/jeronimo_render_proxy/home"

func init() {
	// Register the private render function with the render proxy
	cross(home.Register)(render)
}

func render(string) string {
	return "Rendered by v1"
}

func Render(string) string {
	return "[Home](" + home.URL("") + ")"
}
