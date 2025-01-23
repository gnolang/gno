package testing

import (
	"regexp"
	"time"
)

func X_unixNano() int64 {
	return time.Now().UnixNano()
}

func X_matchString(pat, str string) (result bool, err error) {
	var matchRe *regexp.Regexp
	if matchRe, err = regexp.Compile(pat); err != nil {
		return
	}
	return matchRe.MatchString(str), nil
}
