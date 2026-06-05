package playground

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// maxDecompressedCodeSize caps DEFLATE-decompressed shared code to guard
// against decompression bombs when ?code=...&z are present in query path.
const maxDecompressedCodeSize = 1 << 20 // 1 MiB

const defaultCode = `package main

func Render(path string) string {
	return "Hello, Playground!"
}
`

func (h *Handler) GetPlaygroundView(u *weburl.GnoURL) (int, *components.View) {
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
		Remote:      h.deps.Remote,
		ChainId:     h.deps.ChainId,
		Domain:      h.deps.Domain,
	})
}

func (h *Handler) GetForkView(ctx context.Context, u *weburl.GnoURL) (int, *components.View) {
	pkgPath := u.Path
	files, err := h.deps.Client.ListFiles(ctx, pkgPath)
	if err != nil {
		msg := "unable to list files for fork"
		h.deps.Logger.Warn(msg, "path", pkgPath, "error", err)
		return http.StatusBadRequest, components.StatusErrorComponent(msg)
	}

	var allCode strings.Builder
	for _, fileName := range files {
		if !strings.HasSuffix(fileName, ".gno") && fileName != "gnomod.toml" {
			continue
		}

		body, err := h.deps.Client.File(ctx, pkgPath, fileName)
		if err != nil {
			continue
		}

		if allCode.Len() > 0 {
			allCode.WriteString("\n// --- " + fileName + " ---\n\n")
		} else {
			allCode.WriteString("// --- " + fileName + " ---\n\n")
		}
		allCode.Write(body)
	}

	return http.StatusOK, NewPageView(PlaygroundData{
		InitialCode: allCode.String(),
		ForkFrom:    path.Join(h.deps.Domain, pkgPath),
		Remote:      h.deps.Remote,
		ChainId:     h.deps.ChainId,
		Domain:      h.deps.Domain,
		DefaultFile: u.Query.Get("file"),
	})
}

// EvalHandler returns the http.Handler for POST /_/api/eval.
func (h *Handler) EvalHandler() http.Handler {
	return http.HandlerFunc(h.serveEval)
}

// FuncsHandler returns the http.Handler for GET /_/api/funcs.
func (h *Handler) FuncsHandler() http.Handler {
	return http.HandlerFunc(h.serveFuncs)
}

// evalRequest is the JSON request body for the eval endpoint.
type evalRequest struct {
	PkgPath    string `json:"pkg_path"`
	Expression string `json:"expression"`
}

// evalResponse is the JSON response for the eval endpoint.
type evalResponse struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

// funcsResponse represents the JSON response of the funcs endpoint.
type funcsResponse struct {
	Functions []funcInfo `json:"functions"`
	PkgDoc    string     `json:"pkg_doc,omitempty"`
}

type funcInfo struct {
	Name      string      `json:"name"`
	Doc       string      `json:"doc,omitempty"`
	Signature string      `json:"signature"`
	Params    []paramInfo `json:"params,omitempty"`
	Crossing  bool        `json:"crossing"`
}

type paramInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (h *Handler) serveEval(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.limiter != nil && !h.limiter.allow(clientIP(r)) {
		writeJSON(w, http.StatusTooManyRequests, evalResponse{Error: "rate limit exceeded, please slow down"})
		return
	}

	var req evalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, evalResponse{Error: "invalid request body"})
		return
	}

	if req.PkgPath == "" || req.Expression == "" {
		writeJSON(w, http.StatusBadRequest, evalResponse{Error: "pkg_path and expression are required"})
		return
	}

	// Clean the pkg path.
	pkgPath := strings.TrimPrefix(req.PkgPath, h.deps.Domain+"/")
	pkgPath = strings.TrimPrefix(pkgPath, h.deps.Domain)
	pkgPath = strings.TrimPrefix(pkgPath, "/")

	// Build the qeval data string: "gno.land/r/demo/boards.Render("")".
	data := fmt.Sprintf("%s/%s.%s", h.deps.Domain, pkgPath, req.Expression)

	h.deps.Logger.Debug("playground eval", "data", data)

	start := time.Now()
	result, err := h.deps.Client.Eval(r.Context(), data)
	took := time.Since(start)

	h.deps.Logger.Debug("playground eval result", "took", took, "error", err)

	if err != nil {
		writeJSON(w, http.StatusOK, evalResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, evalResponse{Result: string(result)})
}

func (h *Handler) serveFuncs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pkgPath := r.URL.Query().Get("path")
	if pkgPath == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "path parameter required"})
		return
	}

	jdoc, err := h.deps.Client.Doc(r.Context(), pkgPath)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"error": err.Error()})
		return
	}

	resp := funcsResponse{
		PkgDoc:    jdoc.PackageDoc,
		Functions: make([]funcInfo, 0, len(jdoc.Funcs)),
	}

	for _, fn := range jdoc.Funcs {
		if fn.Type != "" { // skip methods
			continue
		}

		fi := funcInfo{
			Name:      fn.Name,
			Doc:       fn.Doc,
			Signature: fn.Signature,
			Crossing:  fn.Crossing,
			Params:    make([]paramInfo, 0, len(fn.Params)),
		}
		for _, p := range fn.Params {
			fi.Params = append(fi.Params, paramInfo{Name: p.Name, Type: p.Type})
		}

		resp.Functions = append(resp.Functions, fi)
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
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
