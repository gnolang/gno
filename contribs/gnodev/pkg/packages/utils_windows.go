//go:build windows
// +build windows

package packages

import "strings"

// NormalizeFilepathToPath normalize path an unix like path
func NormalizeFilepathToPath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
