package showcase

import (
	"gno.land/p/demo/ui"
	"strings"
)

type command struct {
	name string
	args []string
}

type module struct {
	name     string
	commands []command
}

var modules []module

func Render(path string) string {
	var dom ui.DOM

	// FIXME: replace with p/demo/mux (depends on #882)
	path = strings.TrimSpace(path)
	parts := strings.Split(path, "/")
	switch {
	case path == "":
		dom.Title = "Showcase"
		// dom.Header.Append(ui.Breadcrumb(ui.Link(...)))
		dom.Body.Append(
			ui.Paragraph("Package showcase provides a framework for creating small modules that showcase numerous details and expose various usages."),
			ui.Paragraph("The primary goal is to facilitate easy demonstration and exploration of small features. Modules self-register, expose commands, and demonstrate different aspects. This framework encourages experimentation and creativity, allowing developers to easily add new modules while keeping the main files untouched."),
		)

		output := ""
		list := ui.BulletList()
		for _, module := range modules {
			list.Append(ui.Link(module.name, module.name))
		}
		dom.Body.Append(list)
	}

	return dom.String()
}
