package avlhelpers

import (
	"gno.land/p/nt/avl/v0"
)

// Iterate the keys in-order starting from the given prefix.
// It calls the provided callback function for each key-value pair encountered.
// If the callback returns true, the iteration is stopped.
// The prefix and keys are treated as byte strings, ignoring possible multi-byte Unicode runes.
func IterateByteStringKeysByPrefix(tree avl.ITree, prefix string, cb avl.IterCbFn) {
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

	tree.Iterate(prefix, end, cb)
}

// Get a list of keys starting from the given prefix. Limit the
// number of results to maxResults.
// The prefix and keys are treated as byte strings, ignoring possible multi-byte Unicode runes.
func ListByteStringKeysByPrefix(tree avl.ITree, prefix string, maxResults int) []string {
	result := []string{}
	IterateByteStringKeysByPrefix(tree, prefix, func(key string, value any) bool {
		result = append(result, key)
		if len(result) >= maxResults {
			return true
		}
		return false
	})
	return result
}
