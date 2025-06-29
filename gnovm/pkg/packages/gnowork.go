package packages

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pelletier/go-toml"
)

// NOTE: UNUSED FOR NOW

type Gnowork struct {
	Paths [][2]string

	fpath string
}

// NOTE: can't use a map for Gnowork.Paths as we want to preseve iteration order

func ParseGnowork(fpath string, bz []byte) (*Gnowork, error) {
	var gw Gnowork
	if err := toml.Unmarshal(bz, &gw); err != nil {
		return nil, fmt.Errorf("failed to parse gnowork.toml at %q: %w", fpath, err)
	}
	gw.fpath = fpath
	return &gw, nil
}

func ParseGnoworkAt(fpath string) (*Gnowork, error) {
	bz, err := os.ReadFile(fpath)
	if err != nil {
		return nil, fmt.Errorf("failed to read gnowork.toml at %q: %w", fpath, err)
	}
	return ParseGnowork(fpath, bz)
}

func (gw *Gnowork) PkgLocalPath(pkgPath string) string {
	parts := strings.Split(path.Clean(pkgPath), "/")

	// match last entry first so general case is above
	rev := slices.Clone(gw.Paths)
	slices.Reverse(rev)

	for _, row := range rev {
		localPath, target := row[0], row[1]

		targetParts := strings.Split(path.Clean(target), "/")

		if sliceHasPrefix(parts, targetParts) {
			trail := path.Join(parts[len(targetParts):]...)
			return path.Join(localPath, trail)
		}
	}

	// fallback to using full package path as local path
	return path.Clean(pkgPath)
}

func (gw *Gnowork) PkgDir(pkgPath string) string {
	localPath := gw.PkgLocalPath(pkgPath)
	return filepath.Join(filepath.Dir(gw.fpath), filepath.FromSlash(localPath))
}

func sliceHasPrefix[T comparable](s []T, prefix []T) bool {
	if len(prefix) > len(s) {
		return false
	}
	for i := range prefix {
		if s[i] != prefix[i] {
			return false
		}
	}
	return true
}
