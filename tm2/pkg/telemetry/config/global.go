package config

import (
	"sync/atomic"
)

var (
	globalConfig       Config
	metricsInitialized atomic.Bool
	tracesInitialized  atomic.Bool
)

func SetMetricsInitialized() bool {
	return metricsInitialized.CompareAndSwap(false, true)
}

func SetTracesInitialized() bool {
	return tracesInitialized.CompareAndSwap(false, true)
}

func SetGlobalConfig(config Config) {
	globalConfig = config
}

func GetGlobalConfig() Config {
	return globalConfig
}
