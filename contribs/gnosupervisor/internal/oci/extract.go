package oci

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	whiteoutPrefix = ".wh."
	whiteoutOpaque = ".wh..wh..opq"

	// MaxExtractSize bounds the bytes one layer may write, so that a
	// compressed layer cannot expand without limit.
	MaxExtractSize = 8 << 30
)

// ApplyLayer unpacks the layer file at layerPath (of the given media type)
// into root, keeping only the entries keep accepts, and honouring the
// layer's whiteouts for those entries. keep is asked about directories with
// a trailing slash. Layers must be applied in manifest order.
//
// Only regular files, directories and hard links to files already extracted
// are written. Symbolic links, devices and FIFOs are refused: gnoland images
// carry none in the paths the supervisor keeps, and a symlink is the one
// entry whose target could be chained out of root. Entry names that climb
// out of root are refused too.
func ApplyLayer(root, layerPath, mediaType string, keep func(name string) bool) error {
	f, err := os.Open(layerPath)
	if err != nil {
		return err
	}
	defer f.Close()

	var r io.Reader = f
	switch mediaType {
	case MediaTypeOCILayerGzip, MediaTypeDockerLayerGz:
		gz, err := gzip.NewReader(f)
		if err != nil {
			return fmt.Errorf("layer: %w", err)
		}
		defer gz.Close()
		r = gz
	case MediaTypeOCILayer:
	default:
		return fmt.Errorf("unsupported layer media type %q", mediaType)
	}

	tr := tar.NewReader(r)
	written := int64(0)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("layer: %w", err)
		}
		name, err := cleanName(hdr.Name)
		if err != nil {
			return err
		}
		if name == "" {
			continue
		}
		dir, base := path.Split(name)
		dir = strings.TrimSuffix(dir, "/")

		switch {
		case base == whiteoutOpaque:
			// The directory hides everything lower layers put in it.
			if dir != "" && (keep(dir) || keep(dir+"/")) {
				p := filepath.Join(root, filepath.FromSlash(dir))
				if err := realParents(root, p); err != nil {
					return err
				}
				if err := clearDir(p); err != nil {
					return err
				}
			}
			continue
		case strings.HasPrefix(base, whiteoutPrefix):
			target := path.Join(dir, strings.TrimPrefix(base, whiteoutPrefix))
			// The target may be a file or a whole directory.
			if keep(target) || keep(target+"/") {
				p := filepath.Join(root, filepath.FromSlash(target))
				if err := realParents(root, p); err != nil {
					return err
				}
				if err := os.RemoveAll(p); err != nil {
					return err
				}
			}
			continue
		}

		if hdr.Typeflag == tar.TypeDir {
			name += "/"
		}
		if !keep(name) {
			continue
		}
		dst := filepath.Join(root, filepath.FromSlash(strings.TrimSuffix(name, "/")))
		if hdr.Typeflag == tar.TypeLink {
			src, err := cleanName(hdr.Linkname)
			if err != nil || src == "" || !keep(src) {
				return fmt.Errorf("layer entry %s: hard link to %q, which is not an extracted file", name, hdr.Linkname)
			}
		}
		if hdr.Typeflag == tar.TypeReg {
			if written += hdr.Size; written > MaxExtractSize {
				return fmt.Errorf("layer writes more than %d bytes", MaxExtractSize)
			}
		}
		if err := writeEntry(root, dst, hdr, tr); err != nil {
			return fmt.Errorf("layer entry %s: %w", name, err)
		}
	}
}

func writeEntry(root, dst string, hdr *tar.Header, r io.Reader) error {
	// Symlinks are checked one at a time, so a path through an extracted
	// symlink could still chain its way out of root: write only through real
	// directories.
	if err := realParents(root, dst); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	switch hdr.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(dst, 0o755)
	case tar.TypeReg:
		// A later layer replaces the file; never write through a link.
		if err := os.RemoveAll(dst); err != nil {
			return err
		}
		f, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(hdr.Mode)&0o755)
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, r); err != nil {
			f.Close()
			return err
		}
		return f.Close()
	case tar.TypeLink:
		src, err := cleanName(hdr.Linkname)
		if err != nil || src == "" {
			return fmt.Errorf("hard link to %q", hdr.Linkname)
		}
		srcPath := filepath.Join(root, filepath.FromSlash(src))
		if err := realParents(root, srcPath); err != nil {
			return err
		}
		if fi, err := os.Lstat(srcPath); err != nil || !fi.Mode().IsRegular() {
			return fmt.Errorf("hard link to %q, which is not an extracted regular file", hdr.Linkname)
		}
		if err := os.RemoveAll(dst); err != nil {
			return err
		}
		return os.Link(srcPath, dst)
	default:
		return fmt.Errorf("unsupported entry type %q", string(hdr.Typeflag))
	}
}

// cleanName turns a tar entry name into a slash-separated relative path, or
// refuses it if it climbs out of the root.
func cleanName(name string) (string, error) {
	n := path.Clean("/" + strings.TrimPrefix(name, "./"))
	if strings.Contains(name, "..") {
		for _, part := range strings.Split(name, "/") {
			if part == ".." {
				return "", fmt.Errorf("layer entry %q climbs out of the root", name)
			}
		}
	}
	return strings.TrimPrefix(n, "/"), nil
}

// realParents fails if a directory between root and dst exists as anything
// but a directory.
func realParents(root, dst string) error {
	rel, err := filepath.Rel(root, filepath.Dir(dst))
	if err != nil || rel == "." {
		return err
	}
	p := root
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		p = filepath.Join(p, part)
		fi, err := os.Lstat(p)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			return fmt.Errorf("%s is not a directory", p)
		}
	}
	return nil
}

func clearDir(dir string) error {
	if fi, err := os.Lstat(dir); err != nil || !fi.IsDir() {
		// Missing: nothing to hide. Not a real directory: never follow it.
		return nil
	}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}
