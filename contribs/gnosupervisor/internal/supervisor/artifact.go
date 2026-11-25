package supervisor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/contribs/gnosupervisor/internal/oci"
)

// What the supervisor takes out of a gnoland image: the static binary, and
// the GNOROOT tree it was built with (stdlibs and examples of the same
// commit). See the Dockerfile at the repository root.
const (
	imageBinary  = "usr/bin/gnoland"
	imageGnoRoot = "gnoroot"
	manifestFile = "artifact.json"
)

// keepFromImage selects the binary, the gnoroot tree, and the directories
// above them (so that a later layer's whiteout of one of those directories
// is honoured).
func keepFromImage(name string) bool {
	switch name {
	case imageBinary, "usr/", "usr/bin/":
		return true
	}
	return strings.HasPrefix(name, imageGnoRoot+"/")
}

// Artifact is a prepared release: a gnoland binary and its GNOROOT.
type Artifact struct {
	Version string
	Digest  string
	Binary  string
	GnoRoot string
}

// artifactRecord is written next to an extracted image. Tree is a hash of
// every extracted file, so an artifact damaged on disk is pulled again
// instead of run.
type artifactRecord struct {
	Version  string `json:"version"`
	Image    string `json:"image"`
	Digest   string `json:"digest"`
	Platform string `json:"platform"`
	Tree     string `json:"tree"`
}

// Store prepares artifacts into a cache directory.
type Store struct {
	Dir      string
	Platform string // os/arch
	Client   *oci.Client
	Log      *slog.Logger
}

// Prepare returns the artifact for leg, pulling it if the cache does not
// hold a verified copy. A pull is extracted into a temporary directory and
// renamed into place only once complete, so an interrupted pull leaves
// nothing a later run could mistake for an artifact. Each artifact has its
// own lock file, so supervisors (or a prefetch) sharing the cache never pull
// or remove the same artifact at once.
func (s *Store) Prepare(ctx context.Context, leg Leg) (Artifact, error) {
	digest := *leg.Entry.Image.Digest
	osName, arch, _ := strings.Cut(s.Platform, "/")
	final := filepath.Join(s.Dir, strings.TrimPrefix(digest, "sha256:")+"-"+osName+"-"+arch)

	if err := os.MkdirAll(s.Dir, 0o700); err != nil {
		return Artifact{}, err
	}
	if err := checkPrivateDir(s.Dir); err != nil {
		return Artifact{}, fmt.Errorf("cache directory: %w", err)
	}
	lock, err := lockPath(ctx, final+".lock")
	if err != nil {
		return Artifact{}, err
	}
	defer lock.Close()

	art := Artifact{
		Version: leg.Version,
		Digest:  digest,
		Binary:  filepath.Join(final, "rootfs", filepath.FromSlash(imageBinary)),
		GnoRoot: filepath.Join(final, "rootfs", imageGnoRoot),
	}

	switch err := verifyArtifact(final, leg, s.Platform); {
	case err == nil:
		return art, nil
	case errors.Is(err, fs.ErrNotExist):
	default:
		s.Log.Warn("cached artifact failed verification; pulling it again", "version", leg.Version, "dir", final, "err", err)
		if err := os.RemoveAll(final); err != nil {
			return Artifact{}, err
		}
	}

	tmp, err := os.MkdirTemp(s.Dir, ".pull-")
	if err != nil {
		return Artifact{}, err
	}
	defer os.RemoveAll(tmp) // a no-op once renamed

	s.Log.Info("pulling release image", "version", leg.Version, "image", leg.Entry.Image.Ref, "digest", digest, "platform", s.Platform)
	if err := s.pull(ctx, leg, filepath.Join(tmp, "rootfs")); err != nil {
		return Artifact{}, fmt.Errorf("pull %s (%s): %w", leg.Version, digest, err)
	}
	tree, err := treeHash(filepath.Join(tmp, "rootfs"))
	if err != nil {
		return Artifact{}, err
	}
	rec, err := json.MarshalIndent(artifactRecord{
		Version:  leg.Version,
		Image:    leg.Entry.Image.Ref,
		Digest:   digest,
		Platform: s.Platform,
		Tree:     tree,
	}, "", "  ")
	if err != nil {
		return Artifact{}, err
	}
	if err := writeFileSync(filepath.Join(tmp, manifestFile), rec); err != nil {
		return Artifact{}, err
	}
	if err := os.Rename(tmp, final); err != nil {
		return Artifact{}, err
	}
	if err := syncDir(s.Dir); err != nil {
		return Artifact{}, err
	}
	s.Log.Info("release image ready", "version", leg.Version, "dir", final)
	return art, nil
}

func (s *Store) pull(ctx context.Context, leg Leg, rootfs string) error {
	ref, err := oci.ParseRef(leg.Entry.Image.Ref)
	if err != nil {
		return err
	}
	layers, err := s.Client.ResolveManifest(ctx, ref, *leg.Entry.Image.Digest, s.Platform)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(rootfs, 0o755); err != nil {
		return err
	}
	blobDir := filepath.Dir(rootfs)
	for _, l := range layers {
		blob, err := s.Client.FetchBlob(ctx, ref, l, blobDir)
		if err != nil {
			return err
		}
		err = oci.ApplyLayer(rootfs, blob, l.MediaType, keepFromImage)
		os.Remove(blob)
		if err != nil {
			return fmt.Errorf("layer %s: %w", l.Digest, err)
		}
	}

	bin := filepath.Join(rootfs, filepath.FromSlash(imageBinary))
	if fi, err := os.Lstat(bin); err != nil || !fi.Mode().IsRegular() {
		return fmt.Errorf("image has no regular file /%s", imageBinary)
	}
	if err := os.Chmod(bin, 0o755); err != nil {
		return err
	}
	stdlibs := filepath.Join(rootfs, imageGnoRoot, "gnovm", "stdlibs")
	if fi, err := os.Stat(stdlibs); err != nil || !fi.IsDir() {
		return fmt.Errorf("image has no /%s/gnovm/stdlibs directory", imageGnoRoot)
	}
	return nil
}

// verifyArtifact checks a cached artifact against its record and the leg.
// It returns an fs.ErrNotExist error when there is nothing cached.
func verifyArtifact(dir string, leg Leg, platform string) error {
	data, err := os.ReadFile(filepath.Join(dir, manifestFile))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if _, statErr := os.Stat(dir); statErr == nil {
				return errors.New("artifact record is missing")
			}
		}
		return err
	}
	var rec artifactRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return fmt.Errorf("artifact record: %w", err)
	}
	if rec.Digest != *leg.Entry.Image.Digest || rec.Platform != platform {
		return fmt.Errorf("artifact record is for %s on %s", rec.Digest, rec.Platform)
	}
	tree, err := treeHash(filepath.Join(dir, "rootfs"))
	if err != nil {
		return err
	}
	if tree != rec.Tree {
		return fmt.Errorf("extracted files hash to %s, the record says %s", tree, rec.Tree)
	}
	return nil
}

// treeHash hashes every path, type, permission, symlink target and file
// content under root, in lexical order.
func treeHash(root string) (string, error) {
	h := sha256.New()
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		fmt.Fprintf(h, "%s\x00%s\x00", filepath.ToSlash(rel), info.Mode().String())
		switch {
		case info.Mode()&fs.ModeSymlink != 0:
			target, err := os.Readlink(p)
			if err != nil {
				return err
			}
			fmt.Fprintf(h, "%s\x00", target)
		case info.Mode().IsRegular():
			f, err := os.Open(p)
			if err != nil {
				return err
			}
			fh := sha256.New()
			_, err = io.Copy(fh, f)
			f.Close()
			if err != nil {
				return err
			}
			fmt.Fprintf(h, "%x\x00", fh.Sum(nil))
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

func writeFileSync(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}
