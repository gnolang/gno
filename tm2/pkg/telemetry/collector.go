package telemetry

import (
	"time"
)

type Collector interface {
	RecordBroadcastTxTimer(data time.Duration)
	RecordBuildBlockTimer(data time.Duration)
}

func RecordBroadcastTxTimer(data time.Duration) {
	switch {
	case promCollector != nil:
		promCollector.RecordBroadcastTxTimer(data)
		fallthrough
	case opentlmCollector != nil:
		opentlmCollector.RecordBroadcastTxTimer(data)
	}
}

func RecordBuildBlockTimer(data time.Duration) {
	switch {
	case promCollector != nil:
		promCollector.RecordBuildBlockTimer(data)
		fallthrough
	case opentlmCollector != nil:
		opentlmCollector.RecordBuildBlockTimer(data)
	}
}
