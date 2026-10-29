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
	"sync/atomic"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"golang.org/x/sync/errgroup"
)

const (
	// maxDecompressedCodeSize caps DEFLATE-decompressed shared code to guard
	// against decompression bombs when ?code=...&z are present in query path.
	maxDecompressedCodeSize = 1 << 20 // 1 MiB

	// maxParallelFileFetches caps the number of client source files fetch requests.
	maxParallelFileFetches = 8

	// maxCopyCodeSize caps the total fetched source bytes of a package copy to
	// guard against memory exhaustion from large or numerous package files.
	maxCopyCodeSize = 1 << 20 // 1 MiB

	// defaultCode defines the default code displayed in the code editor.
	defaultCode = `package main

func Render(path string) string {
	return "Hello, Playground!"
}
`
)

// GetPlaygroundView serves /_/play: ?from=<pkgpath> opens an editable copy
// of that package, ?code= a shared snippet, and nothing the default code.
func (h *Handler) GetPlaygroundView(ctx context.Context, u *weburl.GnoURL, indexData *components.IndexData) (int, *components.View) {
	if from := u.Query.Get("from"); from != "" {
		return h.copyView(ctx, "/"+RelPath(h.deps.Domain, from), u.Query.Get("file"))
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

// errCopyTooLarge signals that the cumulative fetched package source exceeded
// maxCopyCodeSize and the fetch was aborted.
var errCopyTooLarge = errors.New("package source is too large to open in the playground")

// copyView opens an editable copy of the package at pkgPath, its source files
// concatenated, with file shown first when set.
func (h *Handler) copyView(ctx context.Context, pkgPath, file string) (int, *components.View) {
	files, err := h.deps.Client.ListFiles(ctx, pkgPath)
	if err != nil {
		msg := "unable to list the package files"
		h.deps.Logger.Warn(msg, "path", pkgPath, "error", err)
		return http.StatusBadRequest, components.StatusErrorComponent(msg)
	}

	// Fetch the source files in parallel, writing each into its own slot so
	// the final order matches the order returned by ListFiles.
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(maxParallelFileFetches)

	// Total bounds the cumulative fetched bytes, crossing maxCopyCodeSize aborts
	// the group, which cancels the remaining in-flight and queued fetches.
	// Contents is initialized to the number of files. It can contain nils if some
	// files are not source files. Done to keep file order during content concatenation.
	var total atomic.Int64
	contents := make([][]byte, len(files))
	for i, fileanme := range files {
		if !isSource(fileanme) {
			continue
		}

		g.Go(func() error {
			body, err := h.deps.Client.File(ctx, pkgPath, fileanme)
			if err != nil {
				return err
			}

			if total.Add(int64(len(body))) > maxCopyCodeSize {
				return errCopyTooLarge
			}

			contents[i] = body
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		if errors.Is(err, errCopyTooLarge) {
			msg := err.Error()
			h.deps.Logger.Warn(msg, "path", pkgPath, "limit", maxCopyCodeSize)
			return http.StatusRequestEntityTooLarge, components.StatusErrorComponent(msg)
		}

		msg := "unable to fetch the package files"
		h.deps.Logger.Error(msg, "path", pkgPath, "error", err)
		return http.StatusInternalServerError, components.StatusErrorComponent(msg)
	}

	// Pre-size the builder so concatenation allocates a single buffer
	// instead of growing per file.
	var code strings.Builder
	headerOverhead := len("\n// --- ") + len(" ---\n\n")
	code.Grow(int(total.Load()) + len(files)*headerOverhead)

	// Concatenate the fetched files in the original ListFiles order
	for i, fileanme := range files {
		if !isSource(fileanme) {
			continue
		}

		// First content line has no newline
		if code.Len() > 0 {
			code.WriteString("\n")
		}

		code.WriteString("// --- " + fileanme + " ---\n\n")
		code.Write(contents[i])
		contents[i] = nil // release the body so it can be collected as we go
	}

	return http.StatusOK, NewPageView(PlaygroundData{
		InitialCode: code.String(),
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
