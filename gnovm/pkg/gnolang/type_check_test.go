package gnolang

import "testing"

func TestCheckAssignableTo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		xt         Type
		dt         Type
		autoNative bool
		wantPanic  bool
	}{
		{
			name: "nil to nil",
			xt:   nil,
			dt:   nil,
		},
		{
			name: "nil and interface",
			xt:   nil,
			dt:   &InterfaceType{},
		},
		{
			name: "interface to nil",
			xt:   &InterfaceType{},
			dt:   nil,
		},
		{
			name:      "nil to non-nillable",
			xt:        nil,
			dt:        PrimitiveType(StringKind),
			wantPanic: true,
		},
		{
			name: "interface to interface",
			xt:   &InterfaceType{},
			dt:   &InterfaceType{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("checkAssignableTo() did not panic, want panic")
					}
				}()
			}
			checkAssignableTo(tt.xt, tt.dt, tt.autoNative)
		})
	}
}
