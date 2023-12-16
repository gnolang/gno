package watcher

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	events "github.com/gnolang/gno/contribs/gnodev/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"

	"github.com/fsnotify/fsnotify"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
)

type PackageWatcher struct {
	PackagesUpdate <-chan PackageUpdateList
	Errors         <-chan error

	logger  log.Logger
	watcher *fsnotify.Watcher
	pkgs    []string
	ctx     context.Context
	stop    context.CancelFunc
	emitter events.Emitter
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

func NewPackageWatcher(logger log.Logger, emitter events.Emitter) (*PackageWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("unable to watch files: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	p := &PackageWatcher{
		pkgs:    []string{},
		logger:  logger,
		watcher: watcher,
		ctx:     ctx,
		stop:    cancel,
		emitter: emitter,
	}

	p.startWatching()

	return p, nil
}

func (p *PackageWatcher) Stop() {
	p.stop()
}

func (p *PackageWatcher) AddPackages(pkgs ...gnomod.Pkg) error {
	for _, pkg := range pkgs {
		dir := pkg.Dir

		abs, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("unable to get absolute path of %q: %w", dir, err)
		}

		// Find the correct insertion point using sorting search
		index := sort.Search(len(p.pkgs), func(i int) bool {
			return len(p.pkgs[i]) <= len(dir)
		})

		// Check if the string already exists at the insertion point
		if index < len(p.pkgs) && (p.pkgs)[index] == dir {
			continue // Skip as it's a duplicate
		}

		// Add the pakcage to the watcher
		if err := p.watcher.Add(abs); err != nil {
			return fmt.Errorf("unable to watch %q: %w", pkg.Dir, err)
		}
	}

	return nil
}

func (p *PackageWatcher) startWatching() {
	const timeout = time.Millisecond * 500

	cerrs := make(chan error, 1)

	cwatch := make(chan PackageUpdateList)

	go func() {
		defer close(cerrs)
		defer close(cwatch)

		var debounceTimer <-chan time.Time
		var pathList = []string{}
		var err error

		for err == nil {
			select {
			case <-p.ctx.Done():
				err = p.ctx.Err()
			case watchErr := <-p.watcher.Errors:
				err = fmt.Errorf("watch error: %w", watchErr)
			case <-debounceTimer:
				updates := p.generatePackagesUpdateList(pathList)
				for _, update := range updates {
					p.logger.Error("packages update",
						"pkg", update.Package,
						"files", update.Files,
					)
				}
				panic(fmt.Sprintf("%+v", updates))

				cwatch <- updates

				// Notify that we have some packages update
				p.emitter.Emit(events.NewPackagesUpdateEvent(updates))

				// Reset pathList and debounceTimer
				pathList = []string{}
				debounceTimer = nil
			case evt := <-p.watcher.Events:
				if evt.Op != fsnotify.Write {
					continue
				}

				pathList = append(pathList, evt.Name)
				debounceTimer = time.After(timeout)
			}
		}

		cerrs <- err
	}()

	p.PackagesUpdate = cwatch
	p.Errors = cerrs
}

func (p *PackageWatcher) generatePackagesUpdateList(paths []string) PackageUpdateList {
	pkgsUpdate := []events.PackageUpdate{}

	mpkgs := map[string]*events.PackageUpdate{} // pkg -> update
	for _, path := range paths {
		for _, pkg := range p.pkgs {
			if !strings.HasPrefix(pkg, path) {
				continue
			}

			pkgu, ok := mpkgs[pkg]
			if !ok {
				pkgsUpdate = append(pkgsUpdate, events.PackageUpdate{
					Package: pkg,
					Files:   []string{},
				})
				pkgu = &pkgsUpdate[len(pkgsUpdate)-1]
			}

			if len(pkg) == len(path) {
				continue
			}

			pkgu.Files = append(pkgu.Files, path)
		}
	}

	return pkgsUpdate
}
