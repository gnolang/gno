package uint256

import (
	"testing"
)

type logicOpTest struct {
	x, y, want string
}

func TestAndBitForm(t *testing.T) {
    testCases := []struct {
        name string
        x    Uint
        y    Uint
        want Uint
    }{
        {
            name: "all zeros",
            x:    Uint{arr: [4]uint64{0, 0, 0, 0}},
            y:    Uint{arr: [4]uint64{0, 0, 0, 0}},
            want: Uint{arr: [4]uint64{0, 0, 0, 0}},
        },
        {
            name: "all ones",
            x:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
            y:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
            want: Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
        },
        {
			name: "mixed",
			x:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			y:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			want: Uint{arr: [4]uint64{0, 0, 0, 0}},
		},
		{
			name: "mixed 2",
			x: 	  Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			y:    Uint{arr: [4]uint64{0, 0, 0, 0}},
			want: Uint{arr: [4]uint64{0, 0, 0, 0}},
		},
		{
            name: "one operand zero",
            x:    Uint{arr: [4]uint64{0, 0, 0, 0}},
            y:    Uint{arr: [4]uint64{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}},
            want: Uint{arr: [4]uint64{0, 0, 0, 0}},
        },
        {
            name: "one operand all ones",
            x:    Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
            y:    Uint{arr: [4]uint64{0x5555555555555555, 0xAAAAAAAAAAAAAAAA, 0xFFFFFFFFFFFFFFFF, 0x0000000000000000}},
            want: Uint{arr: [4]uint64{0x5555555555555555, 0xAAAAAAAAAAAAAAAA, 0xFFFFFFFFFFFFFFFF, 0x0000000000000000}},
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
			res := new(Uint).And(&tc.x, &tc.y)
            if *res != tc.want {
                t.Errorf("And(%v, %v) = %v, want %v", tc.x, tc.y, *res, tc.want)
            }
        })
    }
}
