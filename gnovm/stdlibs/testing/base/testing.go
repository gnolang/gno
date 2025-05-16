package base

import (
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func X_unixNano() int64 {
	// only implemented in testing stdlibs
	return 0
}

func X_matchString(pat, str string) (bool, string) {
	panic("only implemented in testing stdlibs")
}

func X_recoverWithStacktrace() (gnolang.TypedValue, string) {
	panic("only available in testing stdlibs")
}
