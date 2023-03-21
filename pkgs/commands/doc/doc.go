// Package doc implements support for documentation of Gno packages and realms,
// in a similar fashion to `go doc`.
// As a reference, the [official implementation] for `go doc` is used.
//
// [official implementation]: https://github.com/golang/go/tree/90dde5dec1126ddf2236730ec57511ced56a512d/src/cmd/doc
package doc

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/multierr"
)

// DocumentOption is used to pass options to the [Documentable].Document.
type DocumentOption func(s *documentOptions)

type documentOptions struct {
	all        bool
	src        bool
	unexported bool
	short      bool
	w          io.Writer
}

// WithShowAll shows all symbols when displaying documentation about a package.
func WithShowAll(b bool) DocumentOption {
	return func(s *documentOptions) { s.all = b }
}

// WithSource shows source when documenting a symbol.
func WithSource(b bool) DocumentOption {
	return func(s *documentOptions) { s.src = b }
}

// WithUnexported shows unexported symbols as well as exported.
func WithUnexported(b bool) DocumentOption {
	return func(s *documentOptions) { s.unexported = b }
}

// WithShort shows a one-line representation for each symbol.
func WithShort(b bool) DocumentOption {
	return func(s *documentOptions) { s.short = b }
}

// WithWriter uses the given writer as an output.
// By default, os.Stdout is used.
func WithWriter(w io.Writer) DocumentOption {
	return func(s *documentOptions) { s.w = w }
}

// Documentable is a package, symbol, or accessible which can be documented.
type Documentable interface {
	Document(...DocumentOption) error
}

type documentable struct {
	Dir
	symbol     string
	accessible string
	pkgData    *pkgData
}

func (d *documentable) Document(opts ...DocumentOption) error {
	o := &documentOptions{w: os.Stdout}
	for _, opt := range opts {
		opt(o)
	}

	var err error
	// pkgData may already be initialised if we already had to look to see
	// if it had the symbol we wanted; otherwise initialise it now.
	if d.pkgData == nil {
		d.pkgData, err = newPkgData(d.Dir, o.unexported)
		if err != nil {
			return err
		}
	}

	astpkg, pkg, err := d.pkgData.docPackage(o)
	if err != nil {
		return err
	}

	// copied from go source - map vars, constants and constructors to their respective types.
	typedValue := make(map[*doc.Value]bool)
	constructor := make(map[*doc.Func]bool)
	for _, typ := range pkg.Types {
		pkg.Consts = append(pkg.Consts, typ.Consts...)
		pkg.Vars = append(pkg.Vars, typ.Vars...)
		pkg.Funcs = append(pkg.Funcs, typ.Funcs...)
		if o.unexported || token.IsExported(typ.Name) {
			for _, value := range typ.Consts {
				typedValue[value] = true
			}
			for _, value := range typ.Vars {
				typedValue[value] = true
			}
			for _, fun := range typ.Funcs {
				// We don't count it as a constructor bound to the type
				// if the type itself is not exported.
				constructor[fun] = true
			}
		}
	}

	pp := &pkgPrinter{
		name:        d.pkgData.name,
		pkg:         astpkg,
		file:        ast.MergePackageFiles(astpkg, 0),
		doc:         pkg,
		typedValue:  typedValue,
		constructor: constructor,
		fs:          d.pkgData.fset,
		opt:         o,
		importPath:  d.importPath,
	}
	pp.buf.pkg = pp

	return d.output(pp)
}

func (d *documentable) output(pp *pkgPrinter) (err error) {
	defer func() {
		pp.flush()
		if err == nil {
			err = pp.err
		}
	}()

	switch {
	case d.symbol == "" && d.accessible == "":
		if pp.opt.all {
			pp.allDoc()
			return
		}
		pp.packageDoc()
	case d.symbol == "" && d.accessible != "":
		d.symbol, d.accessible = d.accessible, ""
		fallthrough
	case d.symbol != "" && d.accessible == "":
		pp.symbolDoc(d.symbol)
	default: // both non-empty
		if pp.methodDoc(d.symbol, d.accessible) {
			return
		}
		if pp.fieldDoc(d.symbol, d.accessible) {
			return
		}
	}

	return
}

// ResolveDocumentable returns a Documentable from the given arguments.
// Refer to the documentation of gnodev doc for the formats accepted (in general
// the same as the go doc command).
func ResolveDocumentable(dirs *Dirs, args []string, unexported bool) (Documentable, error) {
	parsed := parseArgParts(args)
	if parsed == nil {
		return nil, fmt.Errorf("commands/doc: invalid arguments: %v", args)
	}

	var candidates []Dir

	// if we have a candidate package name, search dirs for a dir that matches it.
	// prefer directories whose import path match precisely the package
	if parsed[0].typ&argPkg > 0 {
		if s, err := os.Stat(parsed[0].val); err == nil && s.IsDir() {
			// expand to full path
			absVal, err := filepath.Abs(parsed[0].val)
			if err == nil {
				candidates = dirs.findDir(absVal)
			}
		}
		// first arg is either not a dir, or if it matched a local dir it was not
		// valid (ie. not scanned by dirs). try parsing as a package
		if len(candidates) == 0 {
			candidates = dirs.findPackage(parsed[0].val)
		}
		// easy case: we wanted documentation about a package, and we found one!
		if len(parsed) == 1 && len(candidates) > 0 {
			return &documentable{Dir: candidates[0]}, nil
		}
		if len(candidates) == 0 {
			// there are no candidates.
			// if this can be something other than a package, remove argPkg as an
			// option, otherwise return not found.
			if parsed[0].typ == argPkg {
				return nil, fmt.Errorf("commands/doc: package not found: %q (note: local packages are not yet supported)", parsed[0].val)
			}
			parsed[0].typ &= ^argPkg
		} else {
			// if there are candidates, then the first argument was definitely
			// a package. remove it from parsed so we don't worry about it again.
			parsed = parsed[1:]
		}
	}

	// we also (or only) have a symbol/accessible.
	// search for the symbol through the candidates
	if len(candidates) == 0 {
		// no candidates means local directory here
		wd, err := os.Getwd()
		if err == nil {
			candidates = dirs.findDir(wd)
		}
		if len(candidates) == 0 {
			return nil, fmt.Errorf("commands/doc: local packages not yet supported")
		}
	}

	doc := &documentable{}

	var matchFunc func(s string) bool
	if len(parsed) == 2 {
		// assert that we have <sym> and <acc>
		if parsed[0].typ&argSym == 0 || parsed[1].typ&argAcc == 0 {
			panic(fmt.Errorf("invalid remaining parsed: %+v", parsed))
		}
		doc.symbol = parsed[0].val
		doc.accessible = parsed[1].val
		matchFunc = func(s string) bool { return s == parsed[0].val+"."+parsed[1].val }
	} else {
		switch parsed[0].typ {
		case argSym:
			doc.symbol = parsed[0].val
			matchFunc = func(s string) bool { return s == parsed[0].val }
		case argAcc:
			doc.accessible = parsed[0].val
			matchFunc = func(s string) bool { return strings.HasSuffix(s, "."+parsed[0].val) }
		case argSym | argAcc:
			matchFunc = func(s string) bool {
				switch {
				case s == parsed[0].val:
					doc.symbol = parsed[0].val
					return true
				case strings.HasSuffix(s, "."+parsed[0].val):
					doc.accessible = parsed[0].val
					return true
				}
				return false
			}
		default:
			panic(fmt.Errorf("invalid remaining parsed: %+v", parsed))
		}
	}

	var errs []error
	for _, candidate := range candidates {
		pd, err := newPkgData(candidate, unexported)
		if err != nil {
			// silently ignore errors -
			// likely invalid AST in a file.
			errs = append(errs, err)
			continue
		}
		for _, sym := range pd.symbols {
			if matchFunc(sym) {
				doc.Dir = candidate
				doc.pkgData = pd
				// match found. return this as documentable.
				return doc, multierr.Combine(errs...)
			}
		}
	}
	return nil, multierr.Append(
		fmt.Errorf("commands/doc: could not resolve arguments: %v", parsed),
		multierr.Combine(errs...),
	)
}

// these are used to specify the type of argPart.
const (
	// ie. "crypto/cipher".
	argPkg byte = 1 << iota
	// ie. "TrimSuffix", "Builder"
	argSym
	// ie. "WriteString". method or field of a type ("accessible", from the
	// word "accessor")
	argAcc
)

// argPart contains the value of the argument, together with the type of value
// it could be, through the flags argPkg, argSym and argAcc.
type argPart struct {
	val string
	typ byte
}

func (a argPart) String() string {
	var b strings.Builder
	if a.typ&argPkg != 0 {
		b.WriteString("pkg")
	}
	if a.typ&argSym != 0 {
		if b.Len() != 0 {
			b.WriteByte('|')
		}
		b.WriteString("sym")
	}
	if a.typ&argAcc != 0 {
		if b.Len() != 0 {
			b.WriteByte('|')
		}
		b.WriteString("acc")
	}
	if b.Len() == 0 {
		b.WriteString("inv:")
	} else {
		b.WriteByte(':')
	}
	b.WriteString(a.val)
	return b.String()
}

func parseArgParts(args []string) []argPart {
	parsed := make([]argPart, 0, 3)
	switch len(args) {
	case 0:
		parsed = append(parsed, argPart{val: ".", typ: argPkg})
	case 1:
		// allowed syntaxes (acc is method or field, [] marks optional):
		// <pkg>
		// [<pkg>.]<sym>[.<acc>]
		// [<pkg>.][<sym>.]<acc>
		// if the (part) argument contains a slash, then it is most certainly
		// a pkg.
		// note: pkg can be a relative path. this is mostly problematic for ".." and
		// ".". so we count full stops from the last slash.
		slash := strings.LastIndexByte(args[0], '/')
		if args[0] == "." || args[0] == ".." ||
			(slash != -1 && args[0][slash+1:] == "..") {
			// special handling for common ., .. and ./..
			// these will generally work poorly if you try to use the one-argument
			// syntax to access a symbol/accessible.
			parsed = append(parsed, argPart{val: args[0], typ: argPkg})
			break
		}
		switch strings.Count(args[0][slash+1:], ".") {
		case 0:
			t := argPkg | argSym | argAcc
			if slash != -1 {
				t = argPkg
			}
			parsed = append(parsed, argPart{args[0], t})
		case 1:
			pos := strings.IndexByte(args[0][slash+1:], '.') + slash + 1
			// pkg.sym, pkg.acc, sym.acc
			t1, t2 := argPkg|argSym, argSym|argAcc
			if slash != -1 {
				t1 = argPkg
			} else if token.IsExported(args[0]) {
				// See rationale here:
				// https://github.com/golang/go/blob/90dde5dec1126ddf2236730ec57511ced56a512d/src/cmd/doc/main.go#L265
				t1, t2 = argSym, argAcc
			}
			parsed = append(parsed,
				argPart{args[0][:pos], t1},
				argPart{args[0][pos+1:], t2},
			)
		case 2:
			// pkg.sym.acc
			parts := strings.Split(args[0][slash+1:], ".")
			parsed = append(parsed,
				argPart{args[0][:slash+1] + parts[0], argPkg},
				argPart{parts[1], argSym},
				argPart{parts[2], argAcc},
			)
		default:
			return nil
		}
	case 2:
		// <pkg> <sym>, <pkg <acc>, <pkg> <sym>.<acc>
		parsed = append(parsed, argPart{args[0], argPkg})
		switch strings.Count(args[1], ".") {
		case 0:
			parsed = append(parsed, argPart{args[1], argSym | argAcc})
		case 1:
			pos := strings.IndexByte(args[1], '.')
			parsed = append(parsed,
				argPart{args[1][:pos], argSym},
				argPart{args[1][pos+1:], argAcc},
			)
		default:
			return nil
		}
	default:
		return nil
	}
	return parsed
}
