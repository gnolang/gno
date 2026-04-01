package gnoweb

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/safeurl"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

// XXX: rework this part, this is the original method from previous gnoweb
// handlerStatusJSON returns an http.Handler that serves status information as JSON.
func handlerStatusJSON(logger *slog.Logger, cli *client.RPCClient) http.Handler {
	const qpath = ".app/version"

	queryVersion := func(ctx context.Context) (*abci.ResponseQuery, error) {
		qres, err := cli.ABCIQuery(ctx, qpath, []byte{})
		if err != nil {
			return nil, errors.Wrap(err, "query app version")
		}

		return &qres.Response, nil
	}

	startedAt := time.Now()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var ret struct {
			Gnoland struct {
				Connected bool    `json:"connected"`
				Error     *string `json:"error,omitempty"`
				Height    *int64  `json:"height,omitempty"`
				// processed txs
				// active connections

				Version *string `json:"version,omitempty"`
				// Uptime    *float64 `json:"uptime-seconds,omitempty"`
				// Goarch    *string  `json:"goarch,omitempty"`
				// Goos      *string  `json:"goos,omitempty"`
				// GoVersion *string  `json:"go-version,omitempty"`
				// NumCPU    *int     `json:"num_cpu,omitempty"`
			} `json:"gnoland"`
			Website struct {
				// Version string  `json:"version"`
				Uptime    float64 `json:"uptime-seconds"`
				Goarch    string  `json:"goarch"`
				Goos      string  `json:"goos"`
				GoVersion string  `json:"go-version"`
				NumCPU    int     `json:"num_cpu"`
			} `json:"website"`
		}
		ret.Website.Uptime = time.Since(startedAt).Seconds()
		ret.Website.Goarch = runtime.GOARCH
		ret.Website.Goos = runtime.GOOS
		ret.Website.NumCPU = runtime.NumCPU()
		ret.Website.GoVersion = runtime.Version()

		ret.Gnoland.Connected = true
		res, err := queryVersion(ctx)
		if err != nil {
			ret.Gnoland.Connected = false
			errmsg := err.Error()
			ret.Gnoland.Error = &errmsg
		} else {
			version := string(res.Value)
			ret.Gnoland.Version = &version
			ret.Gnoland.Height = &res.Height
		}

		out, _ := json.MarshalIndent(ret, "", "  ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	})
}

// handlerLivenessJSON checks if the gnoweb service itself is running and responding.
func handlerLivenessJSON(logger *slog.Logger) http.Handler {
	startTime := time.Now()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple liveness check - service is up and running
		uptime := time.Since(startTime)
		logger.Debug("liveness check passed", "uptime", uptime)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	})
}

// handlerReadyJSON checks if gnoweb can communicate with the RPC node and serve clients.
func handlerReadyJSON(logger *slog.Logger, cli *client.RPCClient, domain string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Test basic query functionality by checking if we can query paths
		const testPath = "vm/qpaths?limit=1"
		testData := domain + "/"

		qres, err := cli.ABCIQuery(ctx, testPath, []byte(testData))
		switch {
		case err != nil:
		case qres.Response.Error != nil:
			err = qres.Response.Error
		case len(qres.Response.Data) == 0:
			// Node should have at least some paths available
			err = errors.New("empty response from the node")

		default: // ok
			logger.Debug("readiness check passed", "path", testPath)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"status":"ready"}`)
			return
		}

		// not ready
		logger.Warn("readiness check failed", "error", err, "test_path", testPath)
		http.Error(w, "not ready", http.StatusServiceUnavailable)
	})
}

// getChainID fetches the status endpoint and returns the "network" field
func getChainID(ctx context.Context, cli *client.RPCClient) (string, error) {
	status, err := cli.Status(ctx, nil)
	if err != nil {
		return "", err
	}

	return status.NodeInfo.Network, nil
}

// handlerSafeURLScan handles polling for SafeURL scan status.
// GET /api/safeurl/scan/{scanID} returns the current scan status.
func handlerSafeURLScan(logger *slog.Logger, validator *safeurl.Validator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract scan ID from path: /api/safeurl/scan/{scanID}
		path := strings.TrimPrefix(r.URL.Path, "/api/safeurl/scan/")
		scanID := strings.TrimSuffix(path, "/")
		if scanID == "" {
			http.Error(w, "scan ID required", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		result, err := validator.GetScanStatus(ctx, scanID)
		if err != nil {
			logger.Warn("failed to get scan status", "scan_id", scanID, "error", err)
			http.Error(w, "failed to get scan status", http.StatusInternalServerError)
			return
		}

		if result == nil {
			http.Error(w, "scan not found", http.StatusNotFound)
			return
		}

		// Return scan result as JSON
		resp := struct {
			ScanID  string `json:"scanId"`
			URL     string `json:"url"`
			Status  string `json:"status"`
			Verdict string `json:"verdict,omitempty"`
		}{
			ScanID:  result.ScanID,
			URL:     result.URL,
			Status:  result.Status.String(),
			Verdict: result.Verdict,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}
