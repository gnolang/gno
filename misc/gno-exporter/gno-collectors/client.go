package gnoexporter

import (
	"log"
	"net/http"

	rpcClient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type GnoCollector struct {
	client rpcClient.Client

	opts GnoCollectorOpts
}

type GnoCollectorOpts struct {
	RPCURL string

	Collectors []Collector
}

func NewGnoCollector(opts GnoCollectorOpts) (*GnoCollector, error) {
	client, err := rpcClient.NewHTTPClient(opts.RPCURL)
	if err != nil {
		return nil, err
	}

	_, err = client.Status()
	if err != nil {
		return nil, err
	}

	// NOTE(albttx): This could be update to read plugins in .so
	opts.Collectors = append(opts.Collectors,
		AccountCollector{client: client},
		RealmCollector{client: client},
		RDemoUsers{client: client},
		TM2Collector{client: client},
	)

	collector := &GnoCollector{
		client: client,
		opts:   opts,
	}
	return collector, nil
}

func (c GnoCollector) GetClient() rpcClient.Client {
	return c.client
}

func (c GnoCollector) AddCollectors(collectors ...Collector) {
	c.opts.Collectors = append(c.opts.Collectors, collectors...)
}

func (c GnoCollector) Start(addr string) error {
	for _, collector := range c.opts.Collectors {
		log.Println("Setting collector: ", collector.Pattern())
		http.HandleFunc(collector.Pattern(), collector.Collect())
	}

	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(addr, nil)
}
