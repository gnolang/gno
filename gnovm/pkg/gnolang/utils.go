package gnolang

import (
	"slices"
	"strings"
)

func contains(list []string, item string) bool {
	return slices.Contains(list, item)
}

func endsWith(item string, suffixes []string) bool {
	for _, i := range suffixes {
		if strings.HasSuffix(item, i) {
			return true
		}
	}
	return false
}
