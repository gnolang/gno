package storage

import (
	"gno.land/p/openzeppelin/storage"
)

var st = storage.NewStorage()

func Execute(cur realm, shardID string, execFn storage.ExecFn) any {
	return st.Execute(shardID, execFn)
}
