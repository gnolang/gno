package gnoexporter

import (
	"log"
	"net/http"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	rpcClient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RealmCollector is complient with the Collector interface
type RealmCollector struct {
	client rpcClient.Client
}

func (c RealmCollector) Pattern() string {
	return "/metrics/realm"
}

func (c RealmCollector) Collect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registry := prometheus.NewRegistry()
		path := r.URL.Query().Get("path")

		address := gnolang.DerivePkgAddr(path)

		realmAccountBalanceGauge := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gno_realm_account_balance",
				Help: "Gno realm account balance",
				// ConstLabels: prometheus.Labels,
			}, []string{"realm_path", "address", "denom"})
		registry.MustRegister(realmAccountBalanceGauge)

		realmSequenceBalanceGauge := prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "gno_realm_account_sequence",
				Help: "Gno Realm account sequence",
				// ConstLabels: prometheus.Labels,
			})
		registry.MustRegister(realmSequenceBalanceGauge)

		account, err := getAccount(c.client, address.String())
		if err != nil {
			log.Printf("failed to get account '%s', err: %v", address, err)
			return
		}

		realmAccountBalanceGauge.With(prometheus.Labels{
			"realm_path": path,
			"address":    address.String(),
			"denom":      "ugnot",
		}).Set(float64(account.GetCoins().AmountOf("ugnot")))

		realmSequenceBalanceGauge.Add(float64(account.GetSequence()))

		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}
