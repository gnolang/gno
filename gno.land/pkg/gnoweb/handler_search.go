package gnoweb

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// searchMaxConcurrentQueries bounds concurrent node path-queries from the
// search endpoint. With singleflight coalescing concurrent /search.json hits
// into a single in-flight Paths() call (2 outbound RPCs total), this limit
// only governs the brief leader-burst window where new callers can become
// leaders between back-to-back resolutions. 16 is generous for that case;
// raising it would only enlarge the worst-case burst footprint if
// singleflight ever stops coalescing.
const searchMaxConcurrentQueries = 16

// handlerSearchJSON serves the realm and package path lists as JSON. The list is
// fetched live from the chain; the browser caches it and filters locally, so no
// query reaches the node per keystroke. The response is a single cacheable URL
// shared by all clients, which collapses node load under a browser/edge cache.
func handlerSearchJSON(logger *slog.Logger, dir RealmDirectory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		realms, packages, err := dir.Paths(r.Context())
		if err != nil {
			logger.Error("search: unable to list paths", "error", err)
			http.Error(w, "search unavailable", http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=30")
		_ = json.NewEncoder(w).Encode(struct {
			Realms   []string `json:"realms"`
			Packages []string `json:"packages"`
		}{Realms: realms, Packages: packages})
	})
}
