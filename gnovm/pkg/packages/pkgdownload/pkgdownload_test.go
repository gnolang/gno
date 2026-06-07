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

func TestDownload_NoPartialWriteOnTraversal(t *testing.T) {
	// Ensure that when a malicious file is mixed with legitimate files,
	// no files are written at all (upfront validation prevents partial state).
	//
	//   tmp/modcache/
	//     avl/          <- dst
	//     ufmt/
	//       ufmt.gno    <- should not be overwritten
	tmp := t.TempDir()
	dst := filepath.Join(tmp, "modcache", "avl")
	ufmtFile := filepath.Join(tmp, "modcache", "ufmt", "ufmt.gno")

	os.MkdirAll(filepath.Dir(ufmtFile), 0o744)
	os.WriteFile(ufmtFile, []byte("ORIGINAL"), 0o644)

	fetcher := &maliciousFetcher{
		files: []*std.MemFile{
			{Name: "legit.gno", Body: "package avl"},       // legitimate file first
			{Name: "../ufmt/ufmt.gno", Body: "BACKDOORED"}, // malicious file second
		},
	}

	err := pkgdownload.Download("gno.land/p/demo/avl", dst, fetcher)
	if err == nil {
		t.Fatal("Download should have rejected traversal")
	}

	// The legitimate file must NOT have been written because
	// validation happens upfront before any write.
	legitFile := filepath.Join(dst, "legit.gno")
	if _, err := os.Stat(legitFile); err == nil {
		t.Fatalf("legit.gno should not exist: upfront validation should prevent any writes")
	}

	// The target file must remain untouched.
	data, _ := os.ReadFile(ufmtFile)
	if string(data) != "ORIGINAL" {
		t.Fatalf("ufmt.gno was modified despite error, content: %s", data)
	}
}
