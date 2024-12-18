package home

import (
	"std"
	"strings"

	"gno.land/p/n2p5/chonk"

	"gno.land/r/leon/hof"
	"gno.land/r/n2p5/config"
)

var (
	active  = chonk.New()
	preview = chonk.New()
)

func init() {
	hof.Register()
}

// Add appends a string to the preview Chonk.
func Add(chunk string) {
	assertAdmin()
	preview.Add(chunk)
}

// Flush clears the preview Chonk.
func Flush() {
	assertAdmin()
	preview.Flush()
}

// Promote promotes the preview Chonk to the active Chonk
// and creates a new preview Chonk.
func Promote() {
	assertAdmin()
	active = preview
	preview = chonk.New()
}

// Render returns the contents of the scanner for the active or preview Chonk
// based on the path provided.
func Render(path string) string {
	var result string
	scanner := getScanner(path)
	for scanner.Scan() {
		result += scanner.Text()
	}
	return result
}

// assertAdmin panics if the caller is not an admin as defined in the config realm.
func assertAdmin() {
	caller := std.PrevRealm().Addr()
	if !config.IsAdmin(caller) {
		panic("forbidden: must be admin")
	}
}

// getScanner returns the scanner for the active or preview Chonk based
// on the path provided.
func getScanner(path string) *chonk.Scanner {
	if isPreview(path) {
		return preview.Scanner()
	}
	return active.Scanner()
}

// isPreview returns true if the path prefix is "preview".
func isPreview(path string) bool {
	return strings.HasPrefix(path, "preview")
}
