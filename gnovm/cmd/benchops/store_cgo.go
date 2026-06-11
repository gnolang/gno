//go:build cgo

package main

// The mdbx backend is selectable only in cgo builds.
import _ "github.com/gnolang/gno/tm2/pkg/db/mdbxdb"
