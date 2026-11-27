package diff

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
)

// Options controls one execution. Zero values are replaced by defaults in
// resolve.
type Options struct {
	Timeout     time.Duration
	GasLimit    int64
	MaxAlloc    int64
	OutputLimit int
	RootDir     string
}

const (
	defaultTimeout     = 20 * time.Second
	defaultGasLimit    = 50_000_000
	defaultMaxAlloc    = 64 << 20
	defaultOutputLimit = 256 << 10
)

// forGo fills the fields the gc runner reads. It does not look up GNOROOT.
func (o Options) forGo() Options {
	if o.Timeout <= 0 {
		o.Timeout = defaultTimeout
	}
	if o.OutputLimit <= 0 {
		o.OutputLimit = defaultOutputLimit
	}
	return o
}

func (o Options) resolve() (Options, error) {
	if o.Timeout <= 0 {
		o.Timeout = defaultTimeout
	}
	if o.GasLimit <= 0 {
		o.GasLimit = defaultGasLimit
	}
	if o.MaxAlloc <= 0 {
		o.MaxAlloc = defaultMaxAlloc
	}
	if o.OutputLimit <= 0 {
		o.OutputLimit = defaultOutputLimit
	}
	root, err := findRoot(o.RootDir)
	if err != nil {
		return Options{}, err
	}
	o.RootDir = root
	return o, nil
}

func findRoot(explicit string) (string, error) {
	if explicit != "" {
		if !isGnoRoot(explicit) {
			return "", fmt.Errorf("GNOROOT %s has no gnovm/stdlibs", explicit)
		}
		return explicit, nil
	}
	if env := os.Getenv("GNOROOT"); env != "" {
		if !isGnoRoot(env) {
			return "", fmt.Errorf("GNOROOT %s has no gnovm/stdlibs", env)
		}
		return env, nil
	}
	return guessRoot()
}

var guessedRoot struct {
	sync.Once
	dir string
	err error
}

func guessRoot() (string, error) {
	guessedRoot.Do(func() {
		if dir, err := gnoenv.GuessRootDir(); err == nil && isGnoRoot(dir) {
			guessedRoot.dir = dir
			return
		}
		cwd, err := os.Getwd()
		if err != nil {
			guessedRoot.err = err
			return
		}
		for dir := cwd; ; {
			if isGnoRoot(dir) {
				guessedRoot.dir = dir
				return
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
		guessedRoot.err = errors.New("GNOROOT is unset and the gno checkout was not found; set GNOROOT")
	})
	return guessedRoot.dir, guessedRoot.err
}

func isGnoRoot(dir string) bool {
	st, err := os.Stat(filepath.Join(dir, "gnovm", "stdlibs"))
	return err == nil && st.IsDir()
}
