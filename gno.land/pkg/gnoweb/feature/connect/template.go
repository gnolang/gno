package connect

import (
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

// WalletsTemplate renders the /wallets install page. mustParse panics at init
// on a malformed template — misconfiguration surfaces immediately, not on the
// first request.
var WalletsTemplate = mustParse("renderWallets", "templates/wallets.html")

func mustParse(name string, paths ...string) *template.Template {
	t, err := template.New(name).ParseFS(templateFS, paths...)
	if err != nil {
		panic("connect: parse " + paths[0] + ": " + err.Error())
	}
	return t
}
