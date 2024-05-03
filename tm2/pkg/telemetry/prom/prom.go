package prom

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	MetricsEnabled bool   `toml:"enabled"`
	ListenAddr     string `toml:"listen_addr"`
	Namespace      string `toml:"namespace"`
}

func DefaultConfig() *Config {
	return &Config{
		MetricsEnabled: false,
		ListenAddr:     ":26660",
		Namespace:      "tm2",
	}
}

func Init(cfg *Config) (*Collector, error) {
	c := &Collector{}

	c.broadcastTxTimer = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: cfg.Namespace,
			Name:      "broadcast_tx_timer",
			Help:      "broadcast tx duration",
		})

	c.buildBlockTimer = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: cfg.Namespace,
			Name:      "build_block_timer",
			Help:      "block build duration",
		})

	prometheus.MustRegister(
		c.broadcastTxTimer,
		c.buildBlockTimer,
	)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(cfg.ListenAddr, nil)
	}()

	return c, nil
}
