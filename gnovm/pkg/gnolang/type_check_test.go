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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("checkAssignableTo() did not panic, want panic")
					}
				}()
			}
			checkAssignableTo(nil, tt.xt, tt.dt, tt.autoNative)
		})
	}
}

func TestAsExpr(t *testing.T) {
	t.Parallel()

	compositeLitExpr := &CompositeLitExpr{
		Elts: KeyValueExprs{},
	}
	constType := &constTypeExpr{
		Source: compositeLitExpr,
	}
	compositeLitExpr.Type = constType

	tests := []struct {
		name     string
		baseExpr Expr
		as       Expr
		want     int
	}{
		{
			name:     "seen the value",
			baseExpr: compositeLitExpr,
			as:       &BinaryExpr{},
			want:     0,
		},
		{
			name: "nested found",
			baseExpr: &BinaryExpr{
				Left: &CallExpr{
					Func: &FuncLitExpr{
						Type: FuncTypeExpr{},
					},
					Args: []Expr{
						&NameExpr{},
						&TypeAssertExpr{
							X: &NameExpr{},
							Type: &constTypeExpr{
								Source: &CompositeLitExpr{
									Type: &MaybeNativeTypeExpr{},
									Elts: KeyValueExprs{},
								},
							},
						},
						&MapTypeExpr{
							Key:   &NameExpr{},
							Value: &ConstExpr{},
						},
					},
				},
				Right: &BinaryExpr{
					Left: &NameExpr{},
				},
			},
			as:   &TypeAssertExpr{},
			want: 1,
		},
		{
			name: "nested not found",
			baseExpr: &BinaryExpr{
				Left: &CallExpr{
					Func: &FuncLitExpr{
						Type: FuncTypeExpr{},
					},
					Args: []Expr{
						&NameExpr{},
						&TypeAssertExpr{
							X: &NameExpr{},
							Type: &constTypeExpr{
								Source: &CompositeLitExpr{
									Type: &MaybeNativeTypeExpr{},
									Elts: KeyValueExprs{},
								},
							},
						},
						&MapTypeExpr{
							Key:   &NameExpr{},
							Value: &ConstExpr{},
						},
					},
				},
				Right: &BinaryExpr{
					Left: &NameExpr{},
				},
			},
			as:   &StructTypeExpr{},
			want: 0,
		},
		{
			name: "basic literal found",
			baseExpr: &BasicLitExpr{
				Value: "test",
			},
			as:   &BasicLitExpr{},
			want: 1,
		},
		{
			name: "basic literal not found",
			baseExpr: &BasicLitExpr{
				Value: "test",
			},
			as:   &NameExpr{},
			want: 0,
		},
		{
			name: "unary expression found",
			baseExpr: &UnaryExpr{
				X: &NameExpr{},
			},
			as:   &UnaryExpr{},
			want: 1,
		},
		{
			name: "unary expression not found",
			baseExpr: &UnaryExpr{
				X: &NameExpr{},
			},
			as:   &BinaryExpr{},
			want: 0,
		},
		{
			name: "composite literal found",
			baseExpr: &CompositeLitExpr{
				Type: &MaybeNativeTypeExpr{},
				Elts: KeyValueExprs{
					{
						Key:   &NameExpr{},
						Value: &ConstExpr{},
					},
				},
			},
			as:   &CompositeLitExpr{},
			want: 1,
		},
		{
			name: "composite literal not found",
			baseExpr: &CompositeLitExpr{
				Type: &MaybeNativeTypeExpr{},
				Elts: KeyValueExprs{
					{
						Key:   &NameExpr{},
						Value: &ConstExpr{},
					},
				},
			},
			as:   &FuncLitExpr{},
			want: 0,
		},
		{
			name: "selector expression found",
			baseExpr: &SelectorExpr{
				X:   &NameExpr{},
				Sel: "test",
			},
			as:   &SelectorExpr{},
			want: 1,
		},
		{
			name: "selector expression not found",
			baseExpr: &SelectorExpr{
				X:   &NameExpr{},
				Sel: "test",
			},
			as:   &NameExpr{},
			want: 0,
		},
		{
			name: "index expression found",
			baseExpr: &IndexExpr{
				X:     &NameExpr{},
				Index: &ConstExpr{},
			},
			as:   &IndexExpr{},
			want: 1,
		},
		{
			name: "index expression not found",
			baseExpr: &IndexExpr{
				X:     &NameExpr{},
				Index: &ConstExpr{},
			},
			as:   &BinaryExpr{},
			want: 0,
		},
		{
			name: "slice expression found",
			baseExpr: &SliceExpr{
				X:    &NameExpr{},
				Low:  &ConstExpr{},
				High: &ConstExpr{},
				Max:  &ConstExpr{},
			},
			as:   &SliceExpr{},
			want: 1,
		},
		{
			name: "slice expression not found",
			baseExpr: &SliceExpr{
				X:    &NameExpr{},
				Low:  &ConstExpr{},
				High: &ConstExpr{},
				Max:  &ConstExpr{},
			},
			as:   &BinaryExpr{},
			want: 0,
		},
		{
			name: "star expression found",
			baseExpr: &StarExpr{
				X: &NameExpr{},
			},
			as:   &StarExpr{},
			want: 1,
		},
		{
			name: "star expression not found",
			baseExpr: &StarExpr{
				X: &NameExpr{},
			},
			as:   &BinaryExpr{},
			want: 0,
		},
		{
			name: "ref expression found",
			baseExpr: &RefExpr{
				X: &NameExpr{},
			},
			as:   &RefExpr{},
			want: 1,
		},
		{
			name: "ref expression not found",
			baseExpr: &RefExpr{
				X: &NameExpr{},
			},
			as:   &BinaryExpr{},
			want: 0,
		},
		{
			name: "field type expression found",
			baseExpr: &FieldTypeExpr{
				Type: &NameExpr{},
				Tag:  &ConstExpr{},
			},
			as:   &FieldTypeExpr{},
			want: 1,
		},
		{
			name: "field type expression not found",
			baseExpr: &FieldTypeExpr{
				Type: &NameExpr{},
				Tag:  &ConstExpr{},
			},
			as:   &BinaryExpr{},
			want: 0,
		},
		{
			name: "array type expression found",
			baseExpr: &ArrayTypeExpr{
				Len: &ConstExpr{},
				Elt: &NameExpr{},
			},
			as:   &ArrayTypeExpr{},
			want: 1,
		},
		{
			name: "array type expression not found",
			baseExpr: &ArrayTypeExpr{
				Len: &ConstExpr{},
				Elt: &NameExpr{},
			},
			as:   &BinaryExpr{},
			want: 0,
		},
		{
			name: "slice type expression found",
			baseExpr: &SliceTypeExpr{
				Elt: &NameExpr{},
			},
			as:   &SliceTypeExpr{},
			want: 1,
		},
		{
			name: "slice type expression not found",
			baseExpr: &SliceTypeExpr{
				Elt: &NameExpr{},
			},
			as:   &BinaryExpr{},
			want: 0,
		},
		{
			name:     "interface type expression found",
			baseExpr: &InterfaceTypeExpr{},
			as:       &InterfaceTypeExpr{},
			want:     1,
		},
		{
			name:     "interface type expression not found",
			baseExpr: &InterfaceTypeExpr{},
			as:       &BinaryExpr{},
			want:     0,
		},
		{
			name: "chan type expression found",
			baseExpr: &ChanTypeExpr{
				Value: &NameExpr{},
			},
			as:   &ChanTypeExpr{},
			want: 1,
		},
		{
			name: "chan type expression not found",
			baseExpr: &ChanTypeExpr{
				Value: &NameExpr{},
			},
			as:   &BinaryExpr{},
			want: 0,
		},
		{
			name:     "func type expression found",
			baseExpr: &FuncTypeExpr{},
			as:       &FuncTypeExpr{},
			want:     1,
		},
		{
			name:     "func type expression not found",
			baseExpr: &FuncTypeExpr{},
			as:       &BinaryExpr{},
			want:     0,
		},
		{
			name: "map type expression found",
			baseExpr: &MapTypeExpr{
				Key:   &NameExpr{},
				Value: &ConstExpr{},
			},
			as:   &MapTypeExpr{},
			want: 1,
		},
		{
			name: "map type expression not found",
			baseExpr: &MapTypeExpr{
				Key:   &NameExpr{},
				Value: &ConstExpr{},
			},
			as:   &BinaryExpr{},
			want: 0,
		},
		{
			name:     "struct type expression found",
			baseExpr: &StructTypeExpr{},
			as:       &StructTypeExpr{},
			want:     1,
		},
		{
			name:     "struct type expression not found",
			baseExpr: &StructTypeExpr{},
			as:       &BinaryExpr{},
			want:     0,
		},
		{
			name: "const type expression found",
			baseExpr: &constTypeExpr{
				Source: &CompositeLitExpr{},
			},
			as:   &constTypeExpr{},
			want: 1,
		},
		{
			name: "const type expression not found",
			baseExpr: &constTypeExpr{
				Source: &CompositeLitExpr{},
			},
			as:   &BinaryExpr{},
			want: 0,
		},
		{
			name: "maybe native type expression found",
			baseExpr: &MaybeNativeTypeExpr{
				Type: &NameExpr{},
			},
			as:   &MaybeNativeTypeExpr{},
			want: 1,
		},
		{
			name: "maybe native type expression not found",
			baseExpr: &MaybeNativeTypeExpr{
				Type: &NameExpr{},
			},
			as:   &BinaryExpr{},
			want: 0,
		},
		{
			name: "func lit expression found",
			baseExpr: &FuncLitExpr{
				Type:         FuncTypeExpr{},
				HeapCaptures: NameExprs{{Name: "test"}},
			},
			as:   &FuncLitExpr{},
			want: 1,
		},
		{
			name: "func lit expression not found",
			baseExpr: &FuncLitExpr{
				Type: FuncTypeExpr{},
			},
			as:   &BinaryExpr{},
			want: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := getExpr(tt.baseExpr, tt.as, &seen[Expr]{})
			if len(got) != tt.want {
				t.Errorf("len(getExpr()) = %v, want %v", len(got), tt.want)
			}
		})
	}
}
