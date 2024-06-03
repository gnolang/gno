package watcher

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	"time"

	emitter "github.com/gnolang/gno/contribs/gnodev/pkg/emitter"
	events "github.com/gnolang/gno/contribs/gnodev/pkg/events"

	"github.com/fsnotify/fsnotify"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
)

type PackageWatcher struct {
	PackagesUpdate <-chan PackageUpdateList
	Errors         <-chan error

	ctx  context.Context
	stop context.CancelFunc

	logger  *slog.Logger
	watcher *fsnotify.Watcher
	pkgsDir []string
	emitter emitter.Emitter
}

func NewPackageWatcher(logger *slog.Logger, emitter emitter.Emitter) (*PackageWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("unable to watch files: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	p := &PackageWatcher{
		ctx:     ctx,
		stop:    cancel,
		pkgsDir: []string{},
		logger:  logger,
		watcher: watcher,
		emitter: emitter,
	}

	p.startWatching()

	return p, nil
}

func (p *PackageWatcher) startWatching() {
	const timeout = time.Millisecond * 500 // Debounce interval

	errorsChan := make(chan error, 1)
	pkgsUpdateChan := make(chan PackageUpdateList)

	go func() {
		defer close(errorsChan)
		defer close(pkgsUpdateChan)

		var debounceTimer <-chan time.Time
		pathList := []string{}
		var err error

		for err == nil {
			select {
			case <-p.ctx.Done():
				err = p.ctx.Err()
			case watchErr := <-p.watcher.Errors:
				err = fmt.Errorf("watch error: %w", watchErr)
			case <-debounceTimer:
				// Process and emit package updates after the debounce interval
				updates := p.generatePackagesUpdateList(pathList)
				for _, update := range updates {
					p.logger.Info("packages update",
						"pkg", update.Package,
						"files", update.Files,
					)
				}

				// Send updates
				pkgsUpdateChan <- updates
				p.emitter.Emit(&events.PackagesUpdate{
					Pkgs: updates,
				})

				// Reset the path list and debounce timer
				pathList = []string{}
				debounceTimer = nil
			case evt := <-p.watcher.Events:
				// Only handle write operations
				if evt.Op != fsnotify.Write {
					continue
				}

				pathList = append(pathList, evt.Name)

				// Set up the debounce timer
				debounceTimer = time.After(timeout)
			}
		}

		errorsChan <- err // Send any final error to the channel
	}()

	// Set update channels
	p.PackagesUpdate = pkgsUpdateChan
	p.Errors = errorsChan
}

func (p *PackageWatcher) Stop() {
	p.stop()
}

// AddPackages adds new packages to the watcher.
// Packages are sorted by their length in descending order to facilitate easier
// and more efficient matching with corresponding paths. The longest paths are
// compared first.
func (p *PackageWatcher) AddPackages(pkgs ...gnomod.Pkg) error {
	for _, pkg := range pkgs {
		dir := pkg.Dir

		abs, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("unable to get absolute path of %q: %w", dir, err)
		}

		// Use binary search to find the correct insertion point
		index := sort.Search(len(p.pkgsDir), func(i int) bool {
			return len(p.pkgsDir[i]) <= len(dir) // Longest paths first
		})

		// Check for duplicates
		if index < len(p.pkgsDir) && p.pkgsDir[index] == dir {
			continue // Skip
		}

		// Insert the package
		p.pkgsDir = append(p.pkgsDir[:index], append([]string{abs}, p.pkgsDir[index:]...)...)

		// Add the package to the watcher and handle any errors
		if err := p.watcher.Add(abs); err != nil {
			return fmt.Errorf("unable to watch %q: %w", pkg.Dir, err)
		}
	}

	return nil
}

func (p *PackageWatcher) generatePackagesUpdateList(paths []string) PackageUpdateList {
	pkgsUpdate := []events.PackageUpdate{}

	mpkgs := map[string]*events.PackageUpdate{} // Pkg -> Update
	for _, path := range paths {
		for _, pkg := range p.pkgsDir {
			dirPath := filepath.Dir(path)

			// Check if a package directory contain our path directory
			if !strings.HasPrefix(pkg, dirPath) {
				continue
			}

			if len(pkg) == len(path) {
				continue // Skip if pkg == path
			}

			// Accumulate file updates for each package
			pkgu, ok := mpkgs[pkg]
			if !ok {
				pkgsUpdate = append(pkgsUpdate, events.PackageUpdate{
					Package: pkg,
					Files:   []string{},
				})
				pkgu = &pkgsUpdate[len(pkgsUpdate)-1]
			}

			pkgu.Files = append(pkgu.Files, path)
		}
	}

	return pkgsUpdate
}

type PackageUpdateList []events.PackageUpdate

func (pkgsu PackageUpdateList) PackagesPath() []string {
	pkgs := make([]string, len(pkgsu))
	for i, pkg := range pkgsu {
		pkgs[i] = pkg.Package
	}
	return pkgs
}

func (pkgsu PackageUpdateList) FilesPath() []string {
	files := make([]string, 0)
	for _, pkg := range pkgsu {
		files = append(files, pkg.Files...)
	}
	return files
}
