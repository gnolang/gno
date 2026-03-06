package pkgdownload_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type maliciousFetcher struct {
	files []*std.MemFile
}

func (f *maliciousFetcher) FetchPackage(pkgPath string) ([]*std.MemFile, error) {
	return f.files, nil
}

func TestDownload_RejectsPathTraversal(t *testing.T) {
	// Create:
	//   tmp/modcache/
	//     avl/       <- dst
	//     ufmt/
	//       ufmt.gno <- legitimate cached file
	tmp := t.TempDir()
	dst := filepath.Join(tmp, "modcache", "avl")
	ufmtFile := filepath.Join(tmp, "modcache", "ufmt", "ufmt.gno")

	os.MkdirAll(filepath.Dir(ufmtFile), 0o744)
	os.WriteFile(ufmtFile, []byte("ORIGINAL"), 0o644)

	fetcher := &maliciousFetcher{
		files: []*std.MemFile{
			{Name: "../ufmt/ufmt.gno", Body: "BACKDOORED"},
		},
	}

	err := pkgdownload.Download("gno.land/p/demo/avl", dst, fetcher)
	if err == nil {
		data, _ := os.ReadFile(ufmtFile)
		t.Fatalf("Download should have rejected traversal, ufmt.gno is now: %s", data)
	}

	data, _ := os.ReadFile(ufmtFile)
	if string(data) != "ORIGINAL" {
		t.Fatalf("ufmt.gno was modified despite error, content: %s", data)
	}
}
