package gnomod

import "golang.org/x/mod/modfile"

type DeprecatedModFile struct {
	Draft   bool
	Module  *modfile.Module
	Gno     *modfile.Go
	Replace []*modfile.Replace

	Syntax *modfile.FileSyntax
}

func (d *DeprecatedModFile) Migrate() (*File, error) {
	f := &File{}
	f.Draft = d.Draft
	f.Module = d.Module
	f.Gno = d.Gno
	// voluntarily not migrating, because not used/working the same way:
	// - f.Replace = d.Replace
	// - f.Syntax = d.Syntax
	return f, nil
}
