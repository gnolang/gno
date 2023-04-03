package types

// fingerprint returns the first 6 bytes of a byte slice.
// If the slice is less than 6 bytes, the fingerprint
// contains trailing zeroes.
func fingerprint(slice []byte) []byte {
	fingerprint := make([]byte, 6)
	copy(fingerprint, slice)
	return fingerprint
}
