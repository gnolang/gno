package main

import "strings"

// stringSliceVar is a flag.Value that accumulates repeated string values.
type stringSliceVar struct{ dst *[]string }

func newStringSliceVar(dst *[]string) *stringSliceVar { return &stringSliceVar{dst: dst} }

func (v *stringSliceVar) String() string {
	if v.dst == nil {
		return ""
	}
	return strings.Join(*v.dst, ",")
}

func (v *stringSliceVar) Set(s string) error {
	*v.dst = append(*v.dst, s)
	return nil
}
