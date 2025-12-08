package watcher

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	emitter "github.com/gnolang/gno/contribs/gnodev/pkg/emitter"
	events "github.com/gnolang/gno/contribs/gnodev/pkg/events"
	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"

	"github.com/fsnotify/fsnotify"
)

type PackageWatcher struct {
	PackagesUpdate <-chan PackageUpdateList
	Errors         <-chan error

	ctx  context.Context
	stop context.CancelFunc

	logger  *slog.Logger
	watcher *fsnotify.Watcher
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
		filesList := []string{}
		var err error

		for err == nil {
			select {
			case <-p.ctx.Done():
				err = p.ctx.Err()
			case watchErr := <-p.watcher.Errors:
				err = fmt.Errorf("watch error: %w", watchErr)
			case <-debounceTimer:
				// Process and emit package updates after the debounce interval
				updates := p.generatePackagesUpdateList(filesList)
				for _, update := range updates {
					p.logger.Info("packages update",
						"pkg", update.PackageDir,
						"files", update.Files,
					)
				}

				// Send updates
				pkgsUpdateChan <- updates
				p.emitter.Emit(&events.PackagesUpdate{
					Pkgs: updates,
				})

				// Reset the path list and debounce timer
				filesList = []string{}
				debounceTimer = nil
			case evt := <-p.watcher.Events:
				// Only handle write operations
				if evt.Op != fsnotify.Write {
					continue
				}

				filesList = append(filesList, evt.Name)

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

func (p *PackageWatcher) UpdatePackagesWatch(pkgs ...*packages.Package) {
	watchList := p.watcher.WatchList()

	oldPkgs := make(map[string]struct{}, len(watchList))
	for _, path := range watchList {
		oldPkgs[path] = struct{}{}
	}

	newPkgs := make(map[string]struct{}, len(pkgs))
	for _, pkg := range pkgs {
		if pkg.Kind != packages.PackageKindFS {
			continue
		}

		if pkg.Dir == "" {
			continue
		}

		dir, err := filepath.Abs(pkg.Dir)
		if err != nil {
			p.logger.Error("Unable to get absolute path", "path", pkg.Dir, "error", err)
			continue
		}

		newPkgs[dir] = struct{}{}
	}

	for dir := range oldPkgs {
		if _, exists := newPkgs[dir]; !exists {
			p.watcher.Remove(dir)
			p.logger.Debug("Watcher list: removed", "path", dir)
		}
	}

	for dir := range newPkgs {
		if _, exists := oldPkgs[dir]; !exists {
			p.watcher.Add(dir)
			p.logger.Debug("Watcher list: added", "path", dir)
		}
	}
}

func (p *PackageWatcher) generatePackagesUpdateList(paths []string) PackageUpdateList {
	pkgsUpdate := []events.PackageUpdate{}

	mpkgs := map[string]*events.PackageUpdate{} // Pkg -> Update
	watchList := p.watcher.WatchList()
	for _, file := range paths {
		for _, watchDir := range watchList {
			if len(watchDir) == len(file) {
				continue // Skip if watchDir == file
			}

			// Check if a package directory contain our path directory
			dir := filepath.Dir(file)
			if !strings.HasPrefix(watchDir, dir) {
				continue
			}

			// Accumulate file updates for each package
			pkgu, ok := mpkgs[watchDir]
			if !ok {
				pkgsUpdate = append(pkgsUpdate, events.PackageUpdate{
					PackageDir: watchDir,
					Files:      []string{},
				})
				pkgu = &pkgsUpdate[len(pkgsUpdate)-1]
			}

			pkgu.Files = append(pkgu.Files, file)
		}
	}

	return pkgsUpdate
}

type PackageUpdateList []events.PackageUpdate

func (pkgsu PackageUpdateList) PackagesPath() []string {
	pkgs := make([]string, len(pkgsu))
	for i, pkg := range pkgsu {
		pkgs[i] = pkg.PackageDir
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
