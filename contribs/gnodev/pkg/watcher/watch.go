package watcher

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
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

// debounceInterval batches FS events before emitting a package update.
// Variable rather than const so tests can shrink it.
var debounceInterval = 500 * time.Millisecond

func (p *PackageWatcher) startWatching() {
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
				// React to any content-changing operation. Write covers
				// in-place saves; Create/Rename cover atomic-rename saves
				// (sed -i, most editors); Remove covers file deletion.
				// Chmod-only events (touch, permission churn) are skipped.
				if !evt.Op.Has(fsnotify.Write | fsnotify.Create | fsnotify.Rename | fsnotify.Remove) {
					continue
				}

				filesList = append(filesList, evt.Name)

				// Set up the debounce timer
				debounceTimer = time.After(debounceInterval)
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
		if pkg.Kind != packages.KindFS {
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
	// Watches are per-directory (non-recursive), so a file belongs to the
	// package whose directory directly contains it, or to none.
	watchList := p.watcher.WatchList()
	watchDirs := make(map[string]struct{}, len(watchList))
	for _, d := range watchList {
		watchDirs[d] = struct{}{}
	}

	mpkgs := map[string]*events.PackageUpdate{} // PackageDir -> Update
	order := []string{}                         // first-seen package order
	for _, file := range paths {
		dir := filepath.Dir(file)
		if _, ok := watchDirs[dir]; !ok {
			continue
		}
		pkgu, ok := mpkgs[dir]
		if !ok {
			pkgu = &events.PackageUpdate{PackageDir: dir, Files: []string{}}
			mpkgs[dir] = pkgu
			order = append(order, dir)
		}
		pkgu.Files = append(pkgu.Files, file)
	}

	pkgsUpdate := make(PackageUpdateList, 0, len(order))
	for _, dir := range order {
		pkgsUpdate = append(pkgsUpdate, *mpkgs[dir])
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
