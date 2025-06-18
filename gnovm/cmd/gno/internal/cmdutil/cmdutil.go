package cmdutil

import (
	"fmt"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type ProcessedFileSet struct {
	Pn   *gno.PackageNode
	Fset *gno.FileSet
}

type ProcessedPackage struct {
	Dir    string             // directory
	MPkg   *std.MemPackage    // includes all files
	Normal ProcessedFileSet   // includes all prod (and some *_test.gno) files
	Tests  ProcessedFileSet   // includes all xxx_test *_test.gno integration files
	Ftests []ProcessedFileSet // includes all *_filetest.gno filetest files
}

func (ppkg *ProcessedPackage) AddNormal(pn *gno.PackageNode, fset *gno.FileSet) {
	if ppkg.Normal != (ProcessedFileSet{}) {
		panic("normal processed fileset already set")
	}
	ppkg.Normal = ProcessedFileSet{pn, fset}
}

func (ppkg *ProcessedPackage) AddUnderscoreTests(pn *gno.PackageNode, fset *gno.FileSet) {
	if ppkg.Tests != (ProcessedFileSet{}) {
		panic("_test processed fileset already set")
	}
	ppkg.Tests = ProcessedFileSet{pn, fset}
}

func (ppkg *ProcessedPackage) AddFileTest(pn *gno.PackageNode, fset *gno.FileSet) {
	if len(fset.Files) != 1 {
		panic("filetests must have filesets of length 1")
	}
	fname := fset.Files[0].FileName
	/* NOTE: filetests in tests/files do not end with _filetest.gno.
	if !strings.HasSuffix(string(fname), "_filetest.gno") {
		panic(fmt.Sprintf("expected *_filetest.gno but got %q", fname))
	}
	*/
	for _, ftest := range ppkg.Ftests {
		if ftest.Fset.Files[0].FileName == fname {
			panic(fmt.Sprintf("fileetest with name %q already exists", fname))
		}
	}
	ppkg.Ftests = append(ppkg.Ftests, ProcessedFileSet{pn, fset})
}

func (ppkg *ProcessedPackage) GetFileTest(fname string) ProcessedFileSet {
	if !strings.HasSuffix(fname, "_filetest.gno") {
		panic(fmt.Sprintf("expected *_filetest.gno but got %q", fname))
	}
	for _, ftest := range ppkg.Ftests {
		if ftest.Fset.Files[0].FileName == fname {
			return ftest
		}
	}
	panic(fmt.Sprintf("processedFileSet for filetest %q not found", fname))
}
