package gnomod

import (
	"golang.org/x/mod/modfile"
)

// Parsed gno.mod file.
type File struct {
	Module  *modfile.Module
	Go      *modfile.Go
	Require []*modfile.Require
	Replace []*modfile.Replace

	Syntax *modfile.FileSyntax
}
