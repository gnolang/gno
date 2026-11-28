package playground

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// MaxCopySourceBytes caps the sources a copy shows, which the page carries
// as text.
const MaxCopySourceBytes = 1 << 20

const (
	// maxSharedCodeLength is the longest ?code= value read: Share refuses
	// links over 8,000 characters (MAX_SHARE_URL_LENGTH), so a longer value
	// was made by hand, and falls back to the default code undecoded.
	maxSharedCodeLength = 8_000

	// maxDecompressedCodeSize caps the JSON a ?code=...&z link inflates to.
	// About 8,000 characters of DEFLATE hold some 30 KiB of real source;
	// the margin is for code that compresses well, not for a link whose
	// few bytes would expand into a page of megabytes.
	maxDecompressedCodeSize = 64 << 10

	// defaultCode defines the default code displayed in the code editor.
	defaultCode = `package main

func Render(path string) string {
	return "Hello, Playground!"
}
`
)

// GetPlaygroundView serves /_/play: ?from=<pkgpath> opens an editable copy
// of that package, ?code= a shared snippet, and nothing the default code.
func (h *Handler) GetPlaygroundView(ctx context.Context, u *weburl.GnoURL) (int, *components.View) {
	if from := u.Query.Get("from"); from != "" {
		pkg := ParseFrom(h.deps.Domain, from)
		if pkg == nil {
			return http.StatusBadRequest, components.StatusErrorComponent("invalid package path")
		}
		return h.copyView(ctx, pkg.Path, u.Query.Get("file"))
	}

	files, err := sharedFiles(u.Query)
	if len(files) == 0 {
		files = []File{{Name: "main.gno", Content: defaultCode}}
	}
	data := PlaygroundData{Files: files, Domain: h.deps.Domain}
	if err != nil {
		data.Refused = err.Error()
	}
	return http.StatusOK, NewPageView(data)
}

// Why a ?code= link was not opened, as the page tells the reader.
var (
	errSharedTooLong  = errors.New("the link is longer than Share writes")
	errSharedTooLarge = fmt.Errorf("the link holds more than %d KiB of code", maxDecompressedCodeSize>>10)
	errSharedInvalid  = errors.New("the link's code does not decode")
	errSharedName     = errors.New("the link names a file the playground does not take")
)

// sharedFiles reads ?code=. With &z, it is what Share writes: the base64 of
// raw DEFLATE (RFC 1951) of the JSON list of files, [{"name","content"}],
// each file as it is, with no line of it read as anything. Without &z, it
// is plain text, one main.gno. A link that is too long, does not decode,
// or names a file isValidFileName rejects (or one twice) is refused with
// the reason, for the page to say why it shows the default code instead.
// No code, &z or not, is the default code, with nothing to say.
func sharedFiles(q url.Values) ([]File, error) {
	code := q.Get("code")
	switch {
	case code == "":
		return nil, nil
	case len(code) > maxSharedCodeLength:
		return nil, errSharedTooLong
	case !q.Has("z"):
		return []File{{Name: "main.gno", Content: code}}, nil
	}
	data, err := base64.StdEncoding.DecodeString(code)
	if err != nil {
		return nil, errSharedInvalid
	}
	plain, err := decodeCompressedCode(data)
	if err != nil {
		return nil, err
	}
	var files []File
	if err := json.Unmarshal(plain, &files); err != nil || len(files) == 0 {
		return nil, errSharedInvalid
	}
	seen := make(map[string]bool, len(files))
	for _, f := range files {
		if !isValidFileName(f.Name) || seen[f.Name] {
			return nil, errSharedName
		}
		seen[f.Name] = true
	}
	return files, nil
}

// validFileName is the allowlist of playground-files.ts's isValidFileName,
// which Add file, Share and a copy go by: a Gno source or gnomod.toml, no
// path part, no leading dot or dash, 100 bytes at most (a ustar name).
var validFileName = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]*\.gno$`)

func isValidFileName(name string) bool {
	return len(name) <= 100 && (name == "gnomod.toml" || validFileName.MatchString(name))
}

// Errors a PackageFetcher answers with, so the copy page answers as the
// package's own pages do.
var (
	ErrPackageNotFound = errors.New("package not found")
	ErrPackageTooLarge = errors.New("package source is too large to open in the playground")
	ErrPackageTimeout  = errors.New("the node did not send the package in time")
)

// copyView opens an editable copy of the package at pkgPath, its source
// files handed to the editor one by one, with file shown first when set.
func (h *Handler) copyView(ctx context.Context, pkgPath, file string) (int, *components.View) {
	files, err := h.deps.Packages.PackageFiles(ctx, pkgPath)
	switch {
	case err == nil:
	case errors.Is(err, ErrPackageNotFound):
		return http.StatusNotFound, components.StatusErrorComponent(ErrPackageNotFound.Error())
	case errors.Is(err, ErrPackageTooLarge):
		h.deps.Logger.Warn("package too large to copy", "path", pkgPath)
		return http.StatusRequestEntityTooLarge, components.StatusErrorComponent(ErrPackageTooLarge.Error())
	case errors.Is(err, ErrPackageTimeout):
		return http.StatusRequestTimeout, components.StatusErrorComponent(ErrPackageTimeout.Error())
	default:
		msg := "unable to fetch the package files"
		h.deps.Logger.Error(msg, "path", pkgPath, "error", err)
		return http.StatusInternalServerError, components.StatusErrorComponent(msg)
	}

	from := strings.TrimPrefix(pkgPath, "/")
	return http.StatusOK, NewPageView(PlaygroundData{
		Files:       files,
		From:        from,
		Domain:      h.deps.Domain,
		DefaultFile: file,
		Preview:     canPreview(from, files),
	})
}

// renderDecl matches the Render gnoweb renders a realm with, crossing or not,
// which vm/qeval calls too.
var renderDecl = regexp.MustCompile(`(?m)^func Render\(\s*(?:\w+\s+realm\s*,\s*)?(?:\w+\s+)?string\s*\)\s*(?:string\b|\(\s*(?:\w+\s+)?string\s*\))`)

// canPreview reports whether Run can show what the copy of from renders on
// the chain: from is a realm and one of its files, not a test, declares
// Render. The page says so on the Run button, and the controller drops it
// on the first edit.
func canPreview(from string, files []File) bool {
	return strings.HasPrefix(from, "r/") && slices.ContainsFunc(files, func(f File) bool {
		return !components.ClassifyFile(f.Name).IsTest && renderDecl.MatchString(f.Content)
	})
}

// CopyURL is the playground URL that opens an editable copy of the package at
// pkgPath, showing file first when set.
func CopyURL(pkgPath, file string) string {
	q := url.Values{"from": {strings.TrimPrefix(pkgPath, "/")}}
	if file != "" {
		q.Set("file", file)
	}
	return "/_/play?" + q.Encode()
}

// ParseFrom reads a ?from= value: the bare path of a realm or a pure
// package, with or without the chain domain. It is nil for anything else,
// a file, args, a query, a user page or a listing, which no copy opens.
func ParseFrom(domain, from string) *weburl.GnoURL {
	rel := "/" + RelPath(domain, from)
	u, err := weburl.Parse(rel)
	if err != nil || u.Path != rel || u.IsDir() || !(u.IsRealm() || u.IsPure()) {
		return nil
	}
	return u
}

// RelPath strips the chain domain and slashes from a package path:
// "gno.land/r/demo/foo", "/r/demo/foo" and "r/demo/foo" all give "r/demo/foo".
func RelPath(domain, p string) string {
	// TrimLeft, not TrimPrefix: "///evil.com" must not become "//evil.com",
	// which links would read as another host.
	p = strings.TrimLeft(p, "/")
	p = strings.TrimPrefix(p, domain+"/")
	return strings.TrimLeft(p, "/")
}

// InCopy reports whether a copy opens a package's file: its Gno sources and
// gnomod.toml, without the filetests, of which a package can hold hundreds.
func InCopy(name string) bool {
	if strings.HasSuffix(name, "_filetest.gno") {
		return false
	}
	return strings.HasSuffix(name, ".gno") || name == "gnomod.toml"
}

// decodeCompressedCode decompresses DEFLATE data, up to
// maxDecompressedCodeSize.
func decodeCompressedCode(data []byte) ([]byte, error) {
	zr := flate.NewReader(bytes.NewReader(data))
	defer zr.Close()

	// Read one byte past the ceiling so an over-limit payload is detectable
	// rather than silently truncated.
	plain, err := io.ReadAll(io.LimitReader(zr, maxDecompressedCodeSize+1))
	switch {
	case err != nil:
		return nil, errSharedInvalid
	case len(plain) > maxDecompressedCodeSize:
		return nil, errSharedTooLarge
	}
	return plain, nil
}
