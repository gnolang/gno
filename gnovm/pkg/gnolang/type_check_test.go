package gnolang

import "testing"

func TestCheckAssignableTo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		xt        Type
		dt        Type
		wantError string
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
			dt:        StringType,
			wantError: "cannot use nil as string value",
		},
		{
			name: "interface to interface",
			xt:   &InterfaceType{},
			dt:   &InterfaceType{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := checkAssignableTo(nil, tt.xt, tt.dt)
			if tt.wantError != "" {
				if err.Error() != tt.wantError {
					t.Errorf("checkAssignableTo() returned wrong error: want: %v got: %v", tt.wantError, err.Error())
				}
			} else if err != nil {
				t.Errorf("checkAssignableTo() returned unexpected wrong error: got: %v", err.Error())
			}
		})
	}
}
