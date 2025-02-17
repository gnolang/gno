//go:build !windows
// +build !windows

package packages

// NormalizeFilepathToPath normalize path an unix like path
func NormalizeFilepathToPath(path string) string {
	return path
}
