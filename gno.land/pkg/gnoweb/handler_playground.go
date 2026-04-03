package gnoweb

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// PlaygroundConfig holds playground-specific configuration.
type PlaygroundConfig struct {
	Enabled bool // Whether playground features are enabled
}

// playgroundAPIHandler handles playground JSON API endpoints.
type playgroundAPIHandler struct {
	logger *slog.Logger
	client ClientAdapter
	domain string
	remote string
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

// funcsResponse represents function info returned by the funcs endpoint.
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

// handlerPlaygroundEval creates an HTTP handler for expression evaluation.
func handlerPlaygroundEval(logger *slog.Logger, cli ClientAdapter, domain, remote string) http.Handler {
	h := &playgroundAPIHandler{
		logger: logger,
		client: cli,
		domain: domain,
		remote: remote,
	}
	return http.HandlerFunc(h.serveEval)
}

// handlerPlaygroundFuncs creates an HTTP handler for listing functions.
func handlerPlaygroundFuncs(logger *slog.Logger, cli ClientAdapter) http.Handler {
	h := &playgroundAPIHandler{
		logger: logger,
		client: cli,
	}
	return http.HandlerFunc(h.serveFuncs)
}

func (h *playgroundAPIHandler) serveEval(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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

	// Clean the pkg path
	pkgPath := strings.TrimPrefix(req.PkgPath, h.domain+"/")
	pkgPath = strings.TrimPrefix(pkgPath, h.domain)
	pkgPath = strings.TrimPrefix(pkgPath, "/")

	// Build the qeval data string: "gno.land/r/demo/boards.Render("")"
	data := fmt.Sprintf("%s/%s.%s", h.domain, pkgPath, req.Expression)

	h.logger.Debug("playground eval", "data", data)

	start := time.Now()
	result, err := h.client.Eval(r.Context(), data)
	took := time.Since(start)

	h.logger.Debug("playground eval result", "took", took, "error", err)

	if err != nil {
		writeJSON(w, http.StatusOK, evalResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, evalResponse{Result: string(result)})
}

func (h *playgroundAPIHandler) serveFuncs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pkgPath := r.URL.Query().Get("path")
	if pkgPath == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "path parameter required"})
		return
	}

	jdoc, err := h.client.Doc(r.Context(), pkgPath)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"error": err.Error()})
		return
	}

	resp := funcsResponse{
		PkgDoc: jdoc.PackageDoc,
	}

	for _, fn := range jdoc.Funcs {
		if fn.Type != "" { // Skip methods
			continue
		}

		fi := funcInfo{
			Name:      fn.Name,
			Doc:       fn.Doc,
			Signature: fn.Signature,
			Crossing:  fn.Crossing,
		}

		for _, p := range fn.Params {
			fi.Params = append(fi.Params, paramInfo{
				Name: p.Name,
				Type: p.Type,
			})
		}

		resp.Functions = append(resp.Functions, fi)
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}
