package main

import (
	"fmt"
	"strconv"
)

// helper type to know if a flag is set or not
// vs relying on its zero value
type optionalFlag[T string | int] struct {
	set bool
	v   T
}

func (o *optionalFlag[T]) Value() *T {
	if !o.set {
		return nil
	}

	return &o.v
}

func (o *optionalFlag[T]) String() string {
	if !o.set {
		return ""
	}

	switch v := any(o.v).(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	default:
		return ""
	}
}

func (o *optionalFlag[T]) Set(value string) error {
	switch any(o.v).(type) {
	case string:
		o.v = any(value).(T)
	case int:
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("value not int: %w", err)
		}
		o.v = any(v).(T)
	}

	o.set = true

	return nil
}
