package gnoexporter

import (
	"log"
	"net/http"

	rpcClient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"

	_ "github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// AccountCollector is complient with the Collector interface
type AccountCollector struct {
	client rpcClient.Client
}

func (c AccountCollector) Pattern() string {
	return "/metrics/account"
}

func (c AccountCollector) Collect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registry := prometheus.NewRegistry()
		address := r.URL.Query().Get("address")

		accountBalanceGauge := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gno_account_balance",
				Help: "Gno account balance",
				// ConstLabels: prometheus.Labels,
			}, []string{"address", "denom"})
		registry.MustRegister(accountBalanceGauge)

		sequenceBalanceGauge := prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "gno_account_sequence",
				Help: "Gno account sequence",
				// ConstLabels: prometheus.Labels,
			})
		registry.MustRegister(sequenceBalanceGauge)

		account, err := getAccount(c.client, address)
		if err != nil {
			log.Printf("failed to get account '%s', err: %v", address, err)
			return
		}

		accountBalanceGauge.With(prometheus.Labels{
			"address": address,
			"denom":   "ugnot",
		}).Set(float64(account.GetCoins().AmountOf("ugnot")))

		sequenceBalanceGauge.Add(float64(account.GetSequence()))

		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}
