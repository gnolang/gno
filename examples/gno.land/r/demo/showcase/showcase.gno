package showcase

import (
	"strings"

	"gno.land/p/demo/ui"
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
	dom := ui.DOM{Prefix: "/r/demo/showcase:"}

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
		links := []ui.Link{}
		for _, module := range modules {
			links = append(links, ui.Link{Text: module.name, Path: module.name})
		}
		list := ui.BulletList(links...)
		dom.Body.Append(list)
	default:
		dom.Title = "404"
		dom.Header.Append(ui.Breadcrumb{ui.Link{Text: "Showcase", Path: ""}})
		dom.Body.Append(
			ui.Paragraph("page not found"),
		)
	}

	return dom.String()
}
