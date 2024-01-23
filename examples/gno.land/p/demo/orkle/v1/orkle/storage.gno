package orkle

import "gno.land/p/demo/orkle/v1/feed"

type Storage interface {
	Put(value string)
	GetLatest() feed.Value
	GetHistory() []feed.Value
}
