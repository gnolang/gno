package gnoweb

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime"
	"time"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

// XXX: rework this part, this is the original method from previous gnoweb
// handlerStatusJSON returns an http.Handler that serves status information as JSON.
func handlerStatusJSON(logger *slog.Logger, cli *client.RPCClient) http.Handler {
	const qpath = ".app/version"

	queryVersion := func() (*abci.ResponseQuery, error) {
		qres, err := cli.ABCIQuery(qpath, []byte{})
		if err != nil {
			return nil, errors.Wrap(err, "query app version")
		}

		return &qres.Response, nil
	}

	startedAt := time.Now()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		res, err := queryVersion()
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

// getChainID fetches the status endpoint and returns the "network" field
func getChainID(cli *client.RPCClient) (string, error) {
	status, err := cli.Status()
	if err != nil {
		return "", err
	}

	return status.NodeInfo.Network, nil
}
