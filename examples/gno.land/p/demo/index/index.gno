package index

import (
	"gno.land/p/demo/avl"
)

type Indexer interface {
	GetByIndex(string, string)
	IterateByIndex(string)
}

type LargeIndexer struct {
	data *avl.Tree
}

type SmallIndexer struct {
	data []interface{}
}

func String(db *avl.Tree, indexer func(entry interface{}) string) (*avl.Tree, Indexer) {
	db.Iterate("", "", func(key string, value interface{}) bool {
		indexManager.Set(indexer(key), value)
		return false
	})
}

func Int(db *avl.Tree, indexer func(entry interface{}) int) *avl.Tree {
	//
	return nil, nil
}
