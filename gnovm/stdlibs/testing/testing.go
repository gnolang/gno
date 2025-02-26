package testing

import (
	"errors"
)

func X_unixNano() int64 {
	// only implemented in testing stdlibs
	return 0
}

func X_matchString(pat, str string) (result bool, err error) {
	return false, errors.New("only implemented in testing stdlibs")
}
