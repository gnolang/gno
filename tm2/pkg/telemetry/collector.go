package telemetry

import (
	"reflect"
	"time"
)

type Collector interface {
	RecordBroadcastTxTimer(data time.Duration)
	RecordBuildBlockTimer(data time.Duration)
	RecordBlockIntervalSeconds(data time.Duration)
}

func RecordBroadcastTxTimer(data time.Duration) {
	if promCollector != nil && !reflect.ValueOf(promCollector).IsNil() {
		promCollector.RecordBroadcastTxTimer(data)
	}
	if opentlmCollector != nil && !reflect.ValueOf(opentlmCollector).IsNil() {
		opentlmCollector.RecordBroadcastTxTimer(data)
	}
}

func RecordBuildBlockTimer(data time.Duration) {
	if promCollector != nil && !reflect.ValueOf(promCollector).IsNil() {
		promCollector.RecordBuildBlockTimer(data)
	}
	if opentlmCollector != nil && !reflect.ValueOf(opentlmCollector).IsNil() {
		opentlmCollector.RecordBuildBlockTimer(data)
	}
}

func RecordBlockIntervalSeconds(data time.Duration) {
	if promCollector != nil && !reflect.ValueOf(promCollector).IsNil() {
		promCollector.RecordBlockIntervalSeconds(data)
	}
	if opentlmCollector != nil && !reflect.ValueOf(opentlmCollector).IsNil() {
		opentlmCollector.RecordBlockIntervalSeconds(data)
	}
}
