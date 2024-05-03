package gnoexporter

import (
	"net/http"

	rpcClient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type TM2Collector struct {
	client rpcClient.Client
}

func (c TM2Collector) Pattern() string {
	return "/metrics/tm2"
}

func (c TM2Collector) Collect() http.HandlerFunc {
	handler := otelhttp.WithRouteTag(c.Pattern(), http.HandlerFunc(handlerFunc))
	return handler
	// return func(w http.ResponseWriter, r *http.Request) {
	// }
}
