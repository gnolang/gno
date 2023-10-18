package gnoexporter

import (
	"log"
	"net/http"
	"regexp"
	"strconv"

	rpcClient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RDemoUsers is complient with the Collector interface
type RDemoUsers struct {
	client rpcClient.Client
}

func (c RDemoUsers) Pattern() string {
	return "/metrics/r/demo/users"
}

func (c RDemoUsers) Collect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registry := prometheus.NewRegistry()

		realmSequenceBalanceCounter := prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "gno_r_demo_users",
				Help: "Gno Realm account sequence",
				// ConstLabels: prometheus.Labels,
			})
		registry.MustRegister(realmSequenceBalanceCounter)

		res, err := c.client.ABCIQuery("vm/qeval", []byte("gno.land/r/demo/users\ncounter"))
		if err != nil {
			log.Printf("failed to Query r/demo/users.counter, err: %v", err)
			return
		}

		re := regexp.MustCompile("[0-9]+")
		counterStr := re.FindString(string(res.Response.Data))

		counter, err := strconv.Atoi(counterStr)
		if err != nil {
			log.Printf("failed to strconv: %s, err: %v", res.Response.Data, err)
			return
		}

		realmSequenceBalanceCounter.Add(float64(counter))

		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}
