package main

type diffStatus uint

const (
	MISSING_IN_SRC diffStatus = 1
	MISSING_IN_DST diffStatus = 2
	HAS_DIFF       diffStatus = 3
	NO_DIFF        diffStatus = 4
)

func (status diffStatus) String() string {
	switch status {
	case MISSING_IN_SRC:
		return "missing in src"
	case MISSING_IN_DST:
		return "missing in dst"
	case HAS_DIFF:
		return "files differ"
	case NO_DIFF:
		return "files are equal"
	default:
		return "Unknown"
	}
}
