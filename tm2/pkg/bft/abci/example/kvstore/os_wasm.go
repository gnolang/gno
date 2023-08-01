//go:build js && wasm
// +build js,wasm

package kvstore

import "github.com/gnolang/gno/tm2/pkg/db"

const dbBackend db.BackendType = db.MemDBBackend
