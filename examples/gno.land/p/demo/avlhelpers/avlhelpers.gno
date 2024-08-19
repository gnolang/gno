package avlhelpers

import (
	"gno.land/p/demo/avl"
)

// Get a list of keys starting from the given prefix. Limit the
// number of results to maxResults.
// The prefix and keys are treated as byte strings, ignoring possible multi-byte Unicode runes.
func ListByteStringKeysByPrefix(tree avl.Tree, prefix string, maxResults int) []string {
	end := ""
	n := len(prefix)
	// To make the end of the search, increment the final character ASCII by one.
	for n > 0 {
		if ascii := int(prefix[n-1]); ascii < 0xff {
			end = prefix[0:n-1] + string(ascii+1)
			break
		}

		// The last character is 0xff. Try the previous character.
		n--
	}

	result := []string{}
	tree.Iterate(prefix, end, func(key string, value interface{}) bool {
		result = append(result, key)
		if len(result) >= maxResults {
			return true
		}
		return false
	})
	return result
}
