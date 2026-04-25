package gnoweb

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gnolang/gno/gno.land/pkg/networks"
)

// handlerNetworksJSON serves the canonical gno.land network registry as JSON.
// The payload comes from the embedded gno.land/pkg/networks/networks.json so
// gnoweb and the registry stay in sync with the monorepo.
func handlerNetworksJSON(logger *slog.Logger) http.Handler {
	body := networks.Raw()
	sum := sha256.Sum256(body)
	etag := strconv.Quote(hex.EncodeToString(sum[:]))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Registry rarely changes (testnet rotations are infrequent) but is
		// fetched on hot paths (per-tx wizards, wallets). Cache aggressively;
		// the ETag still lets clients revalidate quickly when it does change.
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Header().Set("ETag", etag)
		// Public registry: allow cross-origin reads from wallets/explorers.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		if _, err := w.Write(body); err != nil {
			logger.Debug("write networks response", "err", err)
		}
	})
}
