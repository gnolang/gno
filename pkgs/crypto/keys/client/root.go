package client

import (
	"github.com/gnolang/gno/pkgs/command"
)

type (
	AppItem = command.AppItem
	AppList = command.AppList
)

var mainApps AppList = []AppItem{
	{broadcastApp, "broadcast", "broadcast a signed document", DefaultBroadcastOptions},
	{queryApp, "query", "make an ABCI query", DefaultQueryOptions},
}
