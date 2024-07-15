package main

import "strings"

func main() {
	tests := []struct {
		name      string
		fn        func()
		wantPanic bool
	}{
		{
			name:      "String",
			wantPanic: false,
			fn: func() {
				var a strings.Builder
				a.WriteByte('x')
				b := a
				_ = b.String() // appease vet
			},
		},
	}
	for _, tt := range tests {
		// XXX remove goroutine usage
		// didPanic := make(chan bool)
		var didPanic bool = false
		// go func() {
		func() {
			defer func() { didPanic = recover() != nil }()
			tt.fn()
		}()
		if got := didPanic; got != tt.wantPanic {
			panic("!!!should not happen")
		}
	}
}
