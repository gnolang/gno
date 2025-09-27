package v1

import (
	"gno.land/p/nt/ufmt"

	"gno.land/r/docs/upgradeable/simpleproxy/proxy"
)

var app = &App{
	version: "v1",
	field:   "hello world!",
}

func init() {
	proxy.SetImplementation(cross, app)
}

type App struct {
	version string
	field   string
}

// var _ App = (*proxy.MyApp)(nil) -> implements common interface

func (app *App) Render() string {
	return ufmt.Sprintf("This is the %s of the App!\n\nField: %s\n\n", app.version, app.field)
}

func (app *App) Do(arg string) {
	app.field = arg
}

func (app *App) Version() string {
	return app.version
}

func Render(_ string) string {
	return app.Render()
}
