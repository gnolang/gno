package testing

import "regexp"

func X_unixNano() int64 {
	// only implemented in testing stdlibs
	return 0
}

func X_matchString(pat, str string) (result bool, err error) {
	if matchRe == nil || matchPat != pat {
		matchPat = pat
		matchRe, err = regexp.Compile(matchPat)
		if err != nil {
			return
		}
	}
	return matchRe.MatchString(str), nil
}

var (
	matchPat string
	matchRe  *regexp.Regexp
)
