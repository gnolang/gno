package cmdutil

import (
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type ProcessedFileSet struct {
	Pn   *gno.PackageNode
	Fset *gno.FileSet
}

type ProcessedPackage struct {
	Dir   string             // directory
	MPkg  *std.MemPackage    // includes all files
	Prod  ProcessedFileSet   // includes all prod files
	Test  ProcessedFileSet   // includes all prod (and some *_test.gno) files
	XTest ProcessedFileSet   // includes all xxx_test *_test.gno integration files
	FTest []ProcessedFileSet // includes all *_filetest.gno filetest files
}

func setProcessedFset(pfs *ProcessedFileSet, pn *gno.PackageNode, fset *gno.FileSet, name string) {
	if *pfs != (ProcessedFileSet{}) {
		panic(name + " processed fileset already set")
	}
	*pfs = ProcessedFileSet{pn, fset}
}

func (ppkg *ProcessedPackage) AddNormal(pn *gno.PackageNode, fset *gno.FileSet) {
	setProcessedFset(&ppkg.Prod, pn, fset, "prod")
}

func (ppkg *ProcessedPackage) AddTest(pn *gno.PackageNode, fset *gno.FileSet) {
	setProcessedFset(&ppkg.Test, pn, fset, "test")
}

func (ppkg *ProcessedPackage) AddUnderscoreTests(pn *gno.PackageNode, fset *gno.FileSet) {
	setProcessedFset(&ppkg.XTest, pn, fset, "_test")
}

func (ppkg *ProcessedPackage) AddFileTest(pn *gno.PackageNode, fset *gno.FileSet) {
	if len(fset.Files) != 1 {
		panic("filetests must have filesets of length 1")
	}
	fname := fset.Files[0].FileName
	// NOTE: filetests can be either legacy `_filetest.gno` at the package
	// root or new-style bare `.gno` under `filetests/`; the latter has no
	// suffix-based signal in the FileName alone, so we don't assert on it.
	for _, ftest := range ppkg.FTest {
		if ftest.Fset.Files[0].FileName == fname {
			panic(fmt.Sprintf("fileetest with name %q already exists", fname))
		}
	}
	ppkg.FTest = append(ppkg.FTest, ProcessedFileSet{pn, fset})
}

func (ppkg *ProcessedPackage) GetFileTest(fname string) ProcessedFileSet {
	for _, ftest := range ppkg.FTest {
		if ftest.Fset.Files[0].FileName == fname {
			return ftest
		}
	}
	panic(fmt.Sprintf("processedFileSet for filetest %q not found", fname))
}
