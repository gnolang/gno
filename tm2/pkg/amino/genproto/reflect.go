package genproto

import (
	"reflect"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	gTimestampType = reflect.TypeOf(timestamppb.Timestamp{})
	gDurationType  = reflect.TypeOf(durationpb.Duration{})
)

// NOTE: do not change this definition.
func isListType(rt reflect.Type) bool {
	return rt.Kind() == reflect.Slice ||
		rt.Kind() == reflect.Array
}
