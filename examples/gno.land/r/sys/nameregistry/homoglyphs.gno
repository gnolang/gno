package nameregistry

import (
	"errors"
	"regexp"

	"gno.land/p/nt/avl/rotree"
)

// Matches all numbers used as suffix.
var reNumbersSuffix = regexp.MustCompile(`[0-9]{1}[a-z0-9]*$`)

var homoglyphs = map[rune]rune{
	'0': 'o',
	'1': 'i',
	'l': 'i',
}

// GenHomoglyphVariants returns all possible name variants for homoglyphs 0=o, 1=i and l=i.
// Original name is not included in the returned variants.
func GenHomoglyphVariants(name string) []string {
	names := [][]rune{[]rune(name)}
	for i, letter := range name {
		// Get letter to replace the homoglyph
		l, ok := homoglyphs[letter]
		if !ok {
			continue
		}

		// Duplicate current variants with the replacement
		for _, v := range names {
			v2 := make([]rune, len(name))
			copy(v2, v)

			// Replace the homoglyph by the actual letter
			v2[i] = l
			names = append(names, v2)
		}
	}

	// Remove the original name from the variants and check if there are variants
	names = names[1:]
	if len(names) == 0 {
		return nil
	}

	variants := make([]string, len(names))
	for i, n := range names {
		variants[i] = string(n)
	}
	return variants
}

// CheckNameSimilarity checks that a name is not similar to other name from a name tree.
// It fails when a similar name exists. Similarity is checked though typosquatting with
// o=0, i=1 and i=l, or by removing the numbers suffix from the name.
func CheckNameSimilarity(name string, names ...rotree.IReadOnlyTree) error {
	variants := []string{name}

	// Remove all numbers from name prefix to avoid cases like "jesus000"
	prefix := reNumbersSuffix.ReplaceAllString(name, "")
	if prefix != name {
		variants = append(variants, prefix)
	}

	variants = append(variants, GenHomoglyphVariants(name)...)
	for i, n := range variants {
		for _, storage := range names {
			if !storage.Has(n) {
				continue
			}

			// First name in list is the one being checked
			if i == 0 {
				return errors.New("name is reserved")
			} else {
				return errors.New("similar name already exists: " + n)
			}
		}
	}
	return nil
}
