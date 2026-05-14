package state

import "testing"

func TestValidateOID(t *testing.T) {
	if err := ValidateOID("0123456789abcdef0123456789ABCDEF01234567:42"); err != nil {
		t.Fatalf("valid OID rejected: %v", err)
	}
	if err := ValidateOID("not-an-oid"); err == nil {
		t.Fatal("invalid OID accepted")
	}
}

func TestValidateTID(t *testing.T) {
	// A Gno TypeID is a human-readable string, not a hash — getTypeID
	// emits these into Inspect links, so each must be accepted.
	good := []string{
		"gno.land/r/demo/stateshowcase.User",
		"gno.land/r/demo/stateshowcase.Org",
		"gno.land/p/demo/avl.Tree",
		"int",
		"string",
		"[]gno.land/r/demo/foo.Bar",
		"*gno.land/r/demo/foo.Bar",
		"map[string]int",
	}
	for _, s := range good {
		if err := ValidateTID(s); err != nil {
			t.Fatalf("valid TID %q rejected: %v", s, err)
		}
	}
	long := make([]byte, MaxStateIDLength+1)
	for i := range long {
		long[i] = 'a'
	}
	bad := []string{
		"",                  // empty
		string(long),        // oversized
		"foo.Bar\ninjected", // control char (log/RPC injection)
		"foo.Bar\x00",       // NUL
		"foo\tBar",          // tab
	}
	for _, s := range bad {
		if err := ValidateTID(s); err == nil {
			t.Fatalf("invalid TID %q accepted", s)
		}
	}
}

func TestValidateFile(t *testing.T) {
	good := []string{"foo.gno", "foo/bar.gno", "a_b-c.gno", "p/q/r.gno"}
	for _, s := range good {
		if err := ValidateFile(s); err != nil {
			t.Fatalf("valid file %q rejected: %v", s, err)
		}
	}
	// Path-traversal hardening: regex's char-class allows `.` + `/` which
	// composes into `..` and `/../` — ADR-004 §Threat model promises
	// path-traversal protection at handler entry, so each variant must be
	// rejected. Adding `.gno` suffix proves the regex alone is insufficient.
	bad := []string{
		"../etc/passwd",
		"../other.gno",
		"../../root.gno",
		"foo/../bar.gno",
		"foo/..",
		"/abs/path.gno",
		"..",
	}
	for _, s := range bad {
		if err := ValidateFile(s); err == nil {
			t.Fatalf("invalid file %q accepted", s)
		}
	}
}

func TestValidateHeight(t *testing.T) {
	if n, err := ValidateHeight(""); err != nil || n != 0 {
		t.Fatalf("empty height: got (%d, %v), want (0, nil)", n, err)
	}
	if n, err := ValidateHeight("12345"); err != nil || n != 12345 {
		t.Fatalf("valid height: got (%d, %v)", n, err)
	}
	if _, err := ValidateHeight("-1"); err == nil {
		t.Fatal("negative height accepted")
	}
}

func TestValidateLine(t *testing.T) {
	if n, err := ValidateLine("42"); err != nil || n != 42 {
		t.Fatalf("valid line: got (%d, %v)", n, err)
	}
	if _, err := ValidateLine("0"); err == nil {
		t.Fatal("zero line accepted")
	}
}
