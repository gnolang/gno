package haystack

import (
	"gno.land/p/n2p5/haystack"
)

var storage = haystack.New()

func Render(path string) string {
	return `
Put a Needle in the Haystack.
`
}

// Add takes a fixed-length hex-encoded needle bytes and adds it to the haystack key-value store.
// If storage encounters an error, it will panic.
func Add(_ realm, needleHex string) {
	err := storage.Add(needleHex)
	if err != nil {
		panic(err)
	}
}

// Get takes a fixed-length hex-encoded needle hash and returns the hex-encoded needle bytes.
// If storage encounters an error, it will panic.
func Get(hashHex string) string {
	needleHex, err := storage.Get(hashHex)
	if err != nil {
		panic(err)
	}
	return needleHex
}
