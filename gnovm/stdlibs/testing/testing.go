package testing

import (
	"errors"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func X_unixNano() int64 {
	// only implemented in testing stdlibs
	return 0
}

func X_matchString(pat, str string) (result bool, err error) {
	return false, errors.New("only implemented in testing stdlibs")
}

func X_recoverWithStacktrace() (gnolang.TypedValue, string) {
	panic("only available in testing stdlibs")
}
