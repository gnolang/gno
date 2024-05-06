package prom

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Collector is complient with the telemetry.Collector interface
type Collector struct {
	broadcastTxTimer     prometheus.Histogram
	buildBlockTimer      prometheus.Histogram
	blockIntervalSeconds prometheus.Histogram
}

func (c *Collector) RecordBroadcastTxTimer(data time.Duration) {
	c.broadcastTxTimer.Observe(float64(data.Milliseconds()))
}

func (c *Collector) RecordBuildBlockTimer(data time.Duration) {
	c.buildBlockTimer.Observe(float64(data.Milliseconds()))
}

func (c *Collector) RecordBlockIntervalSeconds(data time.Duration) {
	c.blockIntervalSeconds.Observe(data.Seconds())
}
