package gnoimports

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
)

type Resolver interface {
	Resolve(pkgname string) []*Package
}

type FSResolver struct {
	stdlibs map[string][]*Package
	extlibs map[string][]*Package
}

func NewFSResolver() *FSResolver {
	return &FSResolver{
		stdlibs: map[string][]*Package{},
		extlibs: map[string][]*Package{},
	}
}

func (p *FSResolver) Resolve(pkgname string) []*Package {
	// first stdlibs, then external packages
	return append(p.stdlibs[pkgname], p.extlibs[pkgname]...)
}

func (p *FSResolver) LoadStdPackages(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		files, err := os.ReadDir(path)
		if err != nil {
			return nil
		}

		var gnofiles []string
		for _, file := range files {
			if filepath.Ext(file.Name()) == ".gno" {
				gnofiles = append(gnofiles, filepath.Join(path, file.Name()))
			}
		}
		if len(gnofiles) == 0 {
			return nil
		}

		pkgname, ok := strings.CutPrefix(path, root)
		if !ok {
			return nil
		}
		memPkg := gnolang.ReadMemPackageFromList(gnofiles, strings.TrimPrefix(pkgname, "/"))

		p.stdlibs[memPkg.Name] = append(p.stdlibs[memPkg.Name], &Package{
			MemPackage: *memPkg,
			Dir:        path,
		})
		return nil
	})
}

func (p *FSResolver) LoadPackages(root string) error {
	mods, err := gnomod.ListPkgs(root)
	if err != nil {
		return fmt.Errorf("unable to resolve example folder: %w", err)
	}

	sorted, err := mods.Sort()
	if err != nil {
		return fmt.Errorf("unable to sort pkgs: %w", err)
	}

	for _, modPkg := range sorted.GetNonDraftPkgs() {
		memPkg := gnolang.ReadMemPackage(modPkg.Dir, modPkg.Name)
		if memPkg.Validate() != nil {
			continue
		}

		p.extlibs[memPkg.Name] = append(p.extlibs[memPkg.Name], &Package{
			MemPackage: *memPkg,
			Dir:        modPkg.Dir,
		})
	}

	return nil
}
