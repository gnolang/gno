package main

type diffStatus uint

const (
	missingInSrc diffStatus = iota
	missingInDst
	hasDiff
	noDiff
)

func (status diffStatus) String() string {
	switch status {
	case missingInSrc:
		return "missing in src"
	case missingInDst:
		return "missing in dst"
	case hasDiff:
		return "files differ"
	case noDiff:
		return "files are equal"
	default:
		return "Unknown"
	}
}
