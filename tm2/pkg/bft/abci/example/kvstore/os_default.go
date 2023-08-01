//go:build !wasm && !js
// +build !wasm,!js

package kvstore

import "github.com/gnolang/gno/tm2/pkg/db"

const dbBackend db.BackendType = db.GoLevelDBBackend
