package int256

type logicOpTest struct {
	name string
	x    Int
	y    Int
	want Int
}

//func TestOr(t *testing.T) {
//	tests := []logicOpTest{
//		{
//			name: "all zeroes",
//			// TODO: how initialize int.abs ??
//			x:    Int{abs: *u256.Uint{arr: [4]uint64{0, 0, 0, 0}}, neg: false},
//			y:    Int{abs: *u256.Uint{arr: [4]uint64{0, 0, 0, 0}}, neg: false},
//			want: Int{abs: *u256.Uint{arr: [4]uint64{0, 0, 0, 0}}, neg: false},
//		},
//	}
//
//	for _, tc := range tests {
//		t.Run(tc.name, func(t *testing.T) {
//			res := new(Int).Or(&tc.x, &tc.y)
//			if *res != tc.want {
//				println("hmm wrong")
//			}
//		})
//	}
//}
