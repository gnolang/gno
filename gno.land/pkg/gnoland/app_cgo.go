//go:build cgo

package gnoland

import (
	_ "github.com/gnolang/gno/tm2/pkg/db/lmdbdb"
	_ "github.com/gnolang/gno/tm2/pkg/db/mdbxdb"
)
