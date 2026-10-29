package playground

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

const (
	// maxDecompressedCodeSize caps DEFLATE-decompressed shared code to guard
	// against decompression bombs when ?code=...&z are present in query path.
	maxDecompressedCodeSize = 1 << 20 // 1 MiB

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
			return http.StatusBadRequest, components.StatusErrorComponent("invalid package path: from names a realm or a package, without file or arguments")
		}
		return h.copyView(ctx, pkg.Path, u.Query.Get("file"))
	}

	// If available, read initial source code from a query argument
	initial := u.Query.Get("code")
	if initial != "" {
		if decoded, err := base64.StdEncoding.DecodeString(initial); err == nil {
			// Decompress code when given as DEFLATE compressed data format (RFC 1951)
			if u.Query.Has("z") {
				initial = "" // uses default code if decompress fails
				if plain, ok := decodeCompressedCode(decoded); ok {
					initial = plain
				}
			} else {
				initial = string(decoded)
			}
		}
	}

	if initial == "" {
		initial = defaultCode
	}

	return http.StatusOK, NewPageView(PlaygroundData{
		InitialCode: initial,
		Domain:      h.deps.Domain,
	})
}

// Errors a PackageFetcher answers with, so the copy page answers as the
// package's own pages do.
var (
	ErrPackageNotFound = errors.New("package not found")
	ErrPackageTooLarge = errors.New("package source is too large to open in the playground")
	ErrPackageTimeout  = errors.New("the node did not answer in time")
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

	// A new slice: the fetcher's may be shared with other readers.
	sources := make([]File, 0, len(files))
	for _, f := range files {
		if isSource(f.Name) {
			sources = append(sources, f)
		}
	}

	return http.StatusOK, NewPageView(PlaygroundData{
		Files:       sources,
		From:        strings.TrimPrefix(pkgPath, "/"),
		Domain:      h.deps.Domain,
		DefaultFile: file,
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

// isSource reports whether a file is source code that can be displayed in the code editor.
func isSource(filename string) bool {
	return strings.HasSuffix(filename, ".gno") || filename == "gnomod.toml"
}

// decodeCompressedCode decompresses DEFLATE data.
func decodeCompressedCode(data []byte) (string, bool) {
	zr := flate.NewReader(bytes.NewReader(data))
	defer zr.Close()

	// Read one byte past the ceiling so an over-limit payload is detectable
	// rather than silently truncated. Over-limit is validated after by length.
	plain, err := io.ReadAll(io.LimitReader(zr, maxDecompressedCodeSize+1))
	if err != nil || len(plain) > maxDecompressedCodeSize {
		return "", false
	}
	return string(plain), true
}
