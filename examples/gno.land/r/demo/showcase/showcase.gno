package showcase

import (
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/mux"
	"gno.land/p/demo/ui"
)

const (
	realmPath   = "/r/demo/showcase:"
	base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

// cleanPath sanitizes the input path by replacing characters that are not in the base62Chars with '-'
func cleanPath(path string) string {
	return replaceNotInCutset(path, base62Chars, '-')
}

type command struct {
	name        string
	description string
	callback    func(path string) string
}

func (c *command) path() string { return cleanPath(c.name) }

type module struct {
	name        string
	description string
	commands    []command
}

func (m *module) path() string { return cleanPath(m.name) }

var modules = avl.NewTree() // key: module_name(string), value: commands(module)

func registerModule(mod *module) {
	modules.Set(mod.path(), mod)
}

func RenderCommand(res *mux.ResponseWriter, req *mux.Request) {
	modulevar := cleanPath(req.GetVar("module"))
	commandvar := cleanPath(req.GetVar("command"))

	value, exist := modules.Get(modulevar)
	if !exist {
		res.Write("404: module not found")
		return
	}

	module := value.(*module)

	var cmd *command
	for _, c := range module.commands {
		if c.path() == commandvar {
			cmd = &c
			break
		}
	}

	if cmd == nil {
		res.Write("404: command not found")
		return
	}

	fullpath := module.path() + "/" + cmd.path()
	path := strings.TrimPrefix(req.Path, fullpath)
	path = strings.TrimLeft(path, "/")

	dom := ui.DOM{Prefix: realmPath + fullpath}
	dom.Title = cmd.name
	dom.Body.Append(ui.HR{}, ui.Raw{cmd.callback(path)})

	dom.Footer.Append(ui.HR{}, ui.Link{Text: "[Back]", URL: realmPath + module.path()})

	res.Write(dom.String())
}

func RenderModule(res *mux.ResponseWriter, req *mux.Request) {
	modulevar := cleanPath(req.GetVar("module"))

	value, exist := modules.Get(modulevar)
	if !exist {
		res.Write("404: module not found")
		return
	}

	module := value.(*module)

	dom := ui.DOM{Prefix: realmPath + module.path() + "/"}
	dom.Title = module.name

	if module.description != "" {
		dom.Body.Append(ui.Paragraph(module.description))
	}

	items := ui.BulletList{}
	for _, command := range module.commands {
		items = append(items, ui.Link{Text: command.name, Path: command.path()})
	}

	// add commands list items
	dom.Body.Append(items)

	// add back boutton
	dom.Footer.Append(ui.HR{}, ui.Link{Text: "[Back]", URL: realmPath})

	// write to response
	res.Write(dom.String())
}

func RenderHome(res *mux.ResponseWriter, req *mux.Request) {
	dom := ui.DOM{Prefix: realmPath}
	dom.Title = "Showcase"

	dom.Body.Append(
		ui.Paragraph("Package showcase provides a framework for creating small modules that showcase numerous details and expose various usages."),
		ui.Paragraph(`
The primary goal is to facilitate easy demonstration and exploration of small features.
Modules self-register, expose commands, and demonstrate different aspects.
This framework encourages experimentation and creativity, allowing developers to easily add new modules while keeping the main files untouched.`,
		))

	items := ui.BulletList{}
	modules.Iterate("", "", func(key string, value interface{}) bool {
		module := value.(*module)
		items = append(items, ui.Link{Text: module.name, Path: module.path()})

		// add subitems to menu
		subitems := ui.BulletList{}
		for _, command := range module.commands {
			fullPath := module.path() + "/" + command.path()
			subitems = append(subitems, ui.Link{Text: command.name, Path: fullPath})
		}
		items = append(items, subitems)
	})

	dom.Body.Append(items)

	// write to response
	res.Write(dom.String())
}

func Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", RenderHome)

	// handle module
	router.HandleFunc("{module}", RenderModule)

	// handle module command and following path
	router.HandleFunc("{module}/{command}", RenderCommand)
	router.HandleFunc("{module}/{command}/*", RenderCommand)

	return router.Render(path)
}

func replaceNotInCutset(s, cutset string, repl rune) string {
	var result []rune
	for _, c := range s {
		if !strings.ContainsRune(cutset, c) {
			result = append(result, repl)
		} else {
			result = append(result, c)
		}
	}
	return string(result)
}
