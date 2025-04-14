package config

import (
	"sync/atomic"
)

var (
	globalConfig         Config
	telemetryInitialized atomic.Bool
)

func SetTelemetryInitialized() bool {
	return telemetryInitialized.CompareAndSwap(false, true)
}

func SetGlobalConfig(config Config) {
	globalConfig = config
}

func GetGlobalConfig() Config {
	return globalConfig
}
