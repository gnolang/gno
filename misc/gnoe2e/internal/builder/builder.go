package builder

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
)

// BuildOpts configures a gno binary build.
type BuildOpts struct {
	// Binary is the gno binary to build. Empty defaults to "gnoland".
	// Other accepted value: "gpao".
	Binary string
	GOOS   string // target OS. Empty = runtime.GOOS.
	GOARCH string // target arch. Empty = runtime.GOARCH.
	OutDir string // output directory for the binary. Caller owns cleanup.
}

// Builder produces gnoland binaries.
type Builder interface {
	Build(ctx context.Context, opts BuildOpts) (binaryPath string, err error)
}

// LocalBuilder builds gnoland from the local gno checkout.
type LocalBuilder struct{}

// NewLocalBuilder creates a builder that builds from the local gno checkout.
func NewLocalBuilder() *LocalBuilder {
	return &LocalBuilder{}
}

func (b *LocalBuilder) Build(ctx context.Context, opts BuildOpts) (_ string, retErr error) {
	goos := opts.GOOS
	if goos == "" {
		goos = runtime.GOOS
	}
	goarch := opts.GOARCH
	if goarch == "" {
		goarch = runtime.GOARCH
	}
	binary := opts.Binary
	if binary == "" {
		binary = "gnoland"
	}

	outDir := opts.OutDir
	if outDir == "" {
		dir, err := os.MkdirTemp("", "gnoe2e-build-*")
		if err != nil {
			return "", fmt.Errorf("create build dir: %w", err)
		}
		// A failed build returns no path, so nothing can ever name this
		// directory again. Whoever created it owns removing it.
		defer func() {
			if retErr != nil {
				os.RemoveAll(dir)
			}
		}()
		outDir = dir
	}
	return b.buildFromCheckout(ctx, outDir, goos, goarch, binary)
}

// buildFromCheckout builds from the local gno source tree.
func (b *LocalBuilder) buildFromCheckout(ctx context.Context, outDir, goos, goarch, binary string) (string, error) {
	gnoRoot := gnoenv.RootDir()
	var srcDir string
	switch binary {
	case "gnoland":
		srcDir = filepath.Join(gnoRoot, "gno.land", "cmd", "gnoland")
	case "gpao":
		srcDir = filepath.Join(gnoRoot, "contribs", "gpao")
	default:
		return "", fmt.Errorf("unknown binary %q (accepted: gnoland, gpao)", binary)
	}

	if _, err := os.Stat(srcDir); errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("%s source not found at: %s", binary, srcDir)
	}

	outPath := filepath.Join(outDir, binary)
	cmd := exec.CommandContext(ctx, "go", "build", "-o", outPath, ".")
	cmd.Dir = srcDir
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=0",
		fmt.Sprintf("GOOS=%s", goos),
		fmt.Sprintf("GOARCH=%s", goarch),
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("build %s: %w\n%s", binary, err, out)
	}

	return outPath, nil
}
