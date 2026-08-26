package gnomod

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gnolang/gno/tm2/pkg/std"
)

var ErrNoModFile = errors.New("gnomod.toml doesn't exist")

// ParseDir parses, validates and returns a gno.mod or gnomod.toml file located at dir (does not search parents).
func ParseDir(dir string) (*File, error) {
	ferr := func(err error) (*File, error) {
		return nil, fmt.Errorf("parsing gno.mod/gnomod.toml at %s: %w", dir, err)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return ferr(err)
	}

	for _, fname := range []string{"gnomod.toml", "gno.mod"} {
		fpath := filepath.Join(absDir, fname)
		if _, err := os.Stat(fpath); err == nil {
			b, err := os.ReadFile(fpath)
			if err != nil {
				return ferr(err)
			}
			return ParseBytes(fpath, b)
		}
	}

	return ferr(ErrNoModFile)
}

// ParseFilepath tries to parse gno.mod or gnomod.toml file given the file path.
func ParseFilepath(fpath string) (*File, error) {
	b, err := os.ReadFile(fpath)
	if err != nil {
		return nil, fmt.Errorf("could not read file %q: %w", fpath, err)
	}
	return ParseBytes(fpath, b)
}

// MustParseBytes parses a gnomod.toml or gno.mod file from bytes or panic.
func MustParseBytes(fname string, data []byte) *File {
	mod, err := ParseBytes(fname, data)
	if err != nil {
		panic(fmt.Errorf("parsing bytes %w", err))
	}
	return mod
}

// maxFileSize bounds a gnomod.toml (or deprecated gno.mod) body. The body is
// caller-controlled on the MsgAddPackage/MsgRun path — it rides along in the
// MemPackage, where MemFile.ValidateBasic bounds the file name and nothing else.
//
// The shape of what the decoder will accept is bounded in the decoder itself
// (tm2/pkg/toml: maxNestingDepth, maxKeyDepth), which is the only place it can be
// bounded correctly. Upstream pelletier/go-toml v1 recursed without a depth limit
// — ~400,000 unclosed '[' in under 400KB exhausted Go's 1GB goroutine stack, and
// a stack overflow is a fatal runtime error no recover() can catch, so every node
// processing the message died — and walked a key path once per assignment under
// it, which is quadratic in the body. See tm2/pkg/toml/README.md.
//
// Bounding those axes here instead, from the raw bytes, does not work: a '[' may
// sit in a comment or a string and buy no recursion, a closer may sit in a comment
// and unwind nothing, and a quoted key segment may span newlines and so hide its
// depth from any per-line count. Every byte-level approximation of those limits
// this file has carried was either unsound in one of those directions or rejected
// legitimate content (a '#' comment holding an apostrophe, a replace directive
// holding a Windows path). The decoder counts its own recursion and its own parsed
// key paths, so it is exact.
//
// What remains for this layer is the one bound the decoder cannot infer: how many
// bytes a caller is willing to spend at all. Cost is linear in length once shape
// is bounded, so this is what caps it — the worst 4KB body decodes in ~1.2ms,
// ~586ns/byte across the two decodes a message pays, against the 1250 gas/byte
// the keeper charges over the same bytes (PreprocessGasPerByte, see
// gno.land/pkg/sdk/vm.chargePreprocessGas).
//
// 4KB is far above anything legitimate: the largest gnomod.toml in this repository
// is 123 bytes, and a [[replace]] entry runs ~40 bytes, so 4KB is ~100 of them.
// ParseBytes is the single funnel — ParseDir, ParseFilepath, MustParseBytes,
// ParseMemPackage and ParseCheckGnoMod all reach a decoder through it — so the
// bound also covers genesis, the gno CLI on local files, gnoweb and gnodev, none
// of which have a gas meter to fall back on.
const maxFileSize = 4 << 10 // 4KB

// ParseBytes parses a gnomod.toml or gno.mod file from bytes.
func ParseBytes(fpath string, data []byte) (*File, error) {
	if len(data) > maxFileSize {
		return nil, fmt.Errorf("invalid file at %q: size %d exceeds limit %d", fpath, len(data), maxFileSize)
	}
	f, err := parseBytes(fpath, data)
	if err != nil {
		return nil, err
	}
	if err := f.Validate(); err != nil {
		return nil, err
	}
	return f, nil
}

func parseBytes(fpath string, data []byte) (*File, error) {
	fname := filepath.Base(fpath)

	// gnomod.toml
	switch fname {
	case "gnomod.toml":
		return parseTomlBytes(fname, data)
	case "gno.mod":
		dmf, err := parseDeprecatedDotModBytes(fname, data)
		if err != nil {
			return nil, err
		}
		return dmf.Migrate()
	}

	return nil, fmt.Errorf("invalid file at %q: unknown file type", fpath)
}

// ParseMemPackage parses gnomod.toml or gno.mod from MemPackage.
func ParseMemPackage(mpkg *std.MemPackage) (*File, error) {
	for _, fname := range []string{"gnomod.toml", "gno.mod"} {
		if mf := mpkg.GetFile(fname); mf != nil {
			return ParseBytes(mf.Name, []byte(mf.Body))
		}
	}
	return nil, fmt.Errorf("gnomod.toml not in mem package %s (name=%s): %w", mpkg.Path, mpkg.Name, os.ErrNotExist)
}

// MustParseMemPackage parses gno.mod or gnomod.toml from MemPackage,
// panicking on error.
func MustParseMemPackage(mpkg *std.MemPackage) *File {
	mod, err := ParseMemPackage(mpkg)
	if err != nil {
		panic(fmt.Errorf("parsing mempackage: %w", err))
	}
	return mod
}

var reGnoVersion = regexp.MustCompile(`^([0-9][0-9]*)\.(0|[1-9][0-9]*)(\.(0|[1-9][0-9]*))?([a-z]+[0-9]+)?$`)
