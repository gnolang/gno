package orkle

import "github.com/gnolang/gno/agent2/p/orkle/feed"

type Storage interface {
	Put(value string)
	GetLatest() feed.Value
	GetHistory() []feed.Value
}
