package nameregistry

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
