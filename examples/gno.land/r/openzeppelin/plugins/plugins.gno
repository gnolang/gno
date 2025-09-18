package plugins

import (
	"gno.land/p/openzeppelin/plugins"
	"gno.land/r/openzeppelin/storage"
)

// TODO: Manage user realm and caller realm separately, apply user and package access control.

func Run(cur realm, accountID, id string, args ...any) {
	storage.Execute(cross, accountID, plugins.Run(id, args...))
}

func SetPlugins(cur realm, accountID string, list []*plugins.Plugin) {
	storage.Execute(cross, accountID, plugins.SetPlugins(list))
}
