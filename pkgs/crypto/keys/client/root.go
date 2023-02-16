package client

import (
	"github.com/gnolang/gno/pkgs/command"
)

type (
	AppItem = command.AppItem
	AppList = command.AppList
)

var mainApps AppList = []AppItem{
	{exportApp, "export", "export encrypted private key armor", DefaultExportOptions},
	{importApp, "import", "import encrypted private key armor", DefaultImportOptions},
	{listApp, "list", "list all known keys", DefaultListOptions},
	{signApp, "sign", "sign a document", DefaultSignOptions},
	{verifyApp, "verify", "verify a document signature", DefaultVerifyOptions},
	{broadcastApp, "broadcast", "broadcast a signed document", DefaultBroadcastOptions},
	{queryApp, "query", "make an ABCI query", DefaultQueryOptions},
}
