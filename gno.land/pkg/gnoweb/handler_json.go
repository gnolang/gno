package gnoweb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
)

// handlerQuery handles /query/<pkgpath>.<func>(args...) requests.
// It evaluates expressions via vm/qeval and supports format query param.
// Content negotiation: JSON for Accept: application/json, HTML otherwise.
func handlerQuery(logger *slog.Logger, cli *client.RPCClient, domain string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract the query expression from the URL path
		// URL format: /query/<pkgpath>.<func>(args...)
		path := strings.TrimPrefix(r.URL.Path, "/query/")
		if path == "" || path == r.URL.Path {
			http.Error(w, "missing query expression", http.StatusBadRequest)
			return
		}

		// Parse query params for format
		format := r.URL.Query().Get("format")
		if format == "" {
			format = "json" // Default to JSON for this endpoint
		}

		// Validate format
		switch format {
		case "json", "machine", "string": // ok
		default:
			http.Error(w, fmt.Sprintf("invalid format: %s", format), http.StatusBadRequest)
			return
		}

		// Build the query path with format parameter
		qpath := fmt.Sprintf("vm/qeval?format=%s", url.QueryEscape(format))

		// Query data is the expression (pkgpath.func(args))
		// Prepend domain if path doesn't already include it
		data := path
		if !strings.HasPrefix(path, domain) {
			data = domain + "/" + strings.TrimPrefix(path, "/")
		}

		// Execute the query
		result, err := query(ctx, logger, cli, qpath, []byte(data))
		if err != nil {
			logger.Error("query failed", "path", path, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Content negotiation
		if wantsJSON(r) {
			w.Header().Set("Content-Type", "application/json")
			w.Write(result)
			return
		}

		// HTML response with clickable ObjectIDs
		renderJSONHTML(w, result)
	})
}

// handlerObject handles /object/<objectid> requests.
// It retrieves objects by ObjectID via vm/qobject.
// Content negotiation: JSON for Accept: application/json, HTML otherwise.
func handlerObject(logger *slog.Logger, cli *client.RPCClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract the ObjectID from the URL path
		// URL format: /object/<objectid>
		oid := strings.TrimPrefix(r.URL.Path, "/object/")
		if oid == "" || oid == r.URL.Path {
			http.Error(w, "missing object ID", http.StatusBadRequest)
			return
		}

		// Query the object
		const qpath = "vm/qobject"

		result, err := query(ctx, logger, cli, qpath, []byte(oid))
		if err != nil {
			logger.Error("object query failed", "oid", oid, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Content negotiation
		if wantsJSON(r) {
			w.Header().Set("Content-Type", "application/json")
			w.Write(result)
			return
		}

		// HTML response with clickable ObjectIDs
		renderJSONHTML(w, result)
	})
}

// query executes an ABCI query and returns the result.
func query(ctx context.Context, logger *slog.Logger, cli *client.RPCClient, qpath string, data []byte) ([]byte, error) {
	logger.Debug("querying node", "path", qpath, "data", string(data))

	start := time.Now()
	qres, err := cli.ABCIQuery(ctx, qpath, data)
	took := time.Since(start)

	if err != nil {
		logger.Error("query request failed",
			"path", qpath,
			"data", string(data),
			"error", err,
			"took", took,
		)
		return nil, fmt.Errorf("query failed: %w", err)
	}

	if qres.Response.Error != nil {
		logger.Warn("query response error",
			"path", qpath,
			"data", string(data),
			"error", qres.Response.Error,
			"took", took,
		)
		return nil, qres.Response.Error
	}

	logger.Debug("query succeeded",
		"path", qpath,
		"data", string(data),
		"took", took,
	)

	return qres.Response.Data, nil
}

// wantsJSON checks if the client prefers JSON response.
// Returns true for programmatic clients (curl, APIs), false for browsers.
func wantsJSON(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	// Explicit JSON preference
	if strings.Contains(accept, "application/json") {
		return true
	}
	// Browser: explicit HTML preference (browsers always include text/html)
	if strings.Contains(accept, "text/html") {
		return false
	}
	// Default to JSON for curl (*/*), empty Accept, or other programmatic clients
	return true
}

// renderJSONHTML renders JSON content as HTML with clickable ObjectID links.
// It pretty-prints the JSON and replaces ObjectIDs with links.
func renderJSONHTML(w http.ResponseWriter, jsonData []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Pretty-print the JSON
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, jsonData, "", "  "); err != nil {
		// If pretty-print fails, use raw JSON
		prettyJSON.Reset()
		prettyJSON.Write(jsonData)
	}

	// Escape HTML and add clickable links for ObjectIDs
	linkedJSON := linkifyObjectIDs(prettyJSON.String())

	w.Write([]byte("<pre>" + linkedJSON + "</pre>"))
}

// linkifyObjectIDs replaces ObjectID patterns with clickable links.
// ObjectID format: <40-char hex hash>:<number> (e.g., "a8ada09dee16d791fd406d629fe29bb0ed084a30:4")
func linkifyObjectIDs(json string) string {
	// First escape HTML, then add links
	escaped := escapeHTML(json)

	// ObjectID pattern: 40 hex chars followed by : and a number
	// We need to be careful to only match actual ObjectIDs, not other hex strings
	var result strings.Builder
	result.Grow(len(escaped) * 2) // Pre-allocate for potential link additions

	i := 0
	for i < len(escaped) {
		// Look for potential ObjectID start (40 hex chars)
		if isHexChar(escaped[i]) {
			start := i
			hexCount := 0

			// Count consecutive hex characters
			for i < len(escaped) && isHexChar(escaped[i]) && hexCount < 40 {
				hexCount++
				i++
			}

			// Check if we have exactly 40 hex chars followed by : and digits
			if hexCount == 40 && i < len(escaped) && escaped[i] == ':' {
				colonPos := i
				i++

				// Count digits after colon
				digitStart := i
				for i < len(escaped) && escaped[i] >= '0' && escaped[i] <= '9' {
					i++
				}

				if i > digitStart {
					// We found a valid ObjectID
					oid := escaped[start:i]
					result.WriteString(`<a href="/object/`)
					result.WriteString(oid)
					result.WriteString(`">`)
					result.WriteString(oid)
					result.WriteString(`</a>`)
					continue
				}

				// Not a valid ObjectID, write what we've seen
				result.WriteString(escaped[start : colonPos+1])
				// Continue from digitStart
				i = digitStart
				continue
			}

			// Not an ObjectID, write the hex chars we found
			result.WriteString(escaped[start:i])
			continue
		}

		result.WriteByte(escaped[i])
		i++
	}

	return result.String()
}

// isHexChar returns true if c is a hexadecimal character (0-9, a-f, A-F).
func isHexChar(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// escapeHTML escapes special HTML characters.
func escapeHTML(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '<':
			result.WriteString("&lt;")
		case '>':
			result.WriteString("&gt;")
		case '&':
			result.WriteString("&amp;")
		case '"':
			result.WriteString("&quot;")
		case '\'':
			result.WriteString("&#39;")
		default:
			result.WriteByte(s[i])
		}
	}

	return result.String()
}
