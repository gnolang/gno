// Package all imports all available databases. It is useful mostly in tests.
package all

import (
	_ "github.com/gnolang/gno/tm2/pkg/db/goleveldb"
	_ "github.com/gnolang/gno/tm2/pkg/db/memdb"
)
