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
		want     bool
	}{
		{
			name:     "seen the value",
			baseExpr: compositeLitExpr,
			as:       &BinaryExpr{},
			want:     false,
		},
		{
			name: "nested search true",
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
			want: true,
		},
		{
			name: "nested search false",
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
			want: false,
		},
		{
			name: "basic literal search true",
			baseExpr: &BasicLitExpr{
				Value: "test",
			},
			as:   &BasicLitExpr{},
			want: true,
		},
		{
			name: "basic literal search false",
			baseExpr: &BasicLitExpr{
				Value: "test",
			},
			as:   &NameExpr{},
			want: false,
		},
		{
			name: "unary expression search true",
			baseExpr: &UnaryExpr{
				X: &NameExpr{},
			},
			as:   &UnaryExpr{},
			want: true,
		},
		{
			name: "unary expression search false",
			baseExpr: &UnaryExpr{
				X: &NameExpr{},
			},
			as:   &BinaryExpr{},
			want: false,
		},
		{
			name: "composite literal search true",
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
			want: true,
		},
		{
			name: "composite literal search false",
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
			want: false,
		},
		{
			name: "selector expression search true",
			baseExpr: &SelectorExpr{
				X:   &NameExpr{},
				Sel: "test",
			},
			as:   &SelectorExpr{},
			want: true,
		},
		{
			name: "selector expression search false",
			baseExpr: &SelectorExpr{
				X:   &NameExpr{},
				Sel: "test",
			},
			as:   &NameExpr{},
			want: false,
		},
		{
			name: "index expression search true",
			baseExpr: &IndexExpr{
				X:     &NameExpr{},
				Index: &ConstExpr{},
			},
			as:   &IndexExpr{},
			want: true,
		},
		{
			name: "index expression search false",
			baseExpr: &IndexExpr{
				X:     &NameExpr{},
				Index: &ConstExpr{},
			},
			as:   &BinaryExpr{},
			want: false,
		},
		{
			name: "slice expression search true",
			baseExpr: &SliceExpr{
				X:    &NameExpr{},
				Low:  &ConstExpr{},
				High: &ConstExpr{},
				Max:  &ConstExpr{},
			},
			as:   &SliceExpr{},
			want: true,
		},
		{
			name: "slice expression search false",
			baseExpr: &SliceExpr{
				X:    &NameExpr{},
				Low:  &ConstExpr{},
				High: &ConstExpr{},
				Max:  &ConstExpr{},
			},
			as:   &BinaryExpr{},
			want: false,
		},
		{
			name: "star expression search true",
			baseExpr: &StarExpr{
				X: &NameExpr{},
			},
			as:   &StarExpr{},
			want: true,
		},
		{
			name: "star expression search false",
			baseExpr: &StarExpr{
				X: &NameExpr{},
			},
			as:   &BinaryExpr{},
			want: false,
		},
		{
			name: "ref expression search true",
			baseExpr: &RefExpr{
				X: &NameExpr{},
			},
			as:   &RefExpr{},
			want: true,
		},
		{
			name: "ref expression search false",
			baseExpr: &RefExpr{
				X: &NameExpr{},
			},
			as:   &BinaryExpr{},
			want: false,
		},
		{
			name: "field type expression search true",
			baseExpr: &FieldTypeExpr{
				Type: &NameExpr{},
				Tag:  &ConstExpr{},
			},
			as:   &FieldTypeExpr{},
			want: true,
		},
		{
			name: "field type expression search false",
			baseExpr: &FieldTypeExpr{
				Type: &NameExpr{},
				Tag:  &ConstExpr{},
			},
			as:   &BinaryExpr{},
			want: false,
		},
		{
			name: "array type expression search true",
			baseExpr: &ArrayTypeExpr{
				Len: &ConstExpr{},
				Elt: &NameExpr{},
			},
			as:   &ArrayTypeExpr{},
			want: true,
		},
		{
			name: "array type expression search false",
			baseExpr: &ArrayTypeExpr{
				Len: &ConstExpr{},
				Elt: &NameExpr{},
			},
			as:   &BinaryExpr{},
			want: false,
		},
		{
			name: "slice type expression search true",
			baseExpr: &SliceTypeExpr{
				Elt: &NameExpr{},
			},
			as:   &SliceTypeExpr{},
			want: true,
		},
		{
			name: "slice type expression search false",
			baseExpr: &SliceTypeExpr{
				Elt: &NameExpr{},
			},
			as:   &BinaryExpr{},
			want: false,
		},
		{
			name:     "interface type expression search true",
			baseExpr: &InterfaceTypeExpr{},
			as:       &InterfaceTypeExpr{},
			want:     true,
		},
		{
			name:     "interface type expression search false",
			baseExpr: &InterfaceTypeExpr{},
			as:       &BinaryExpr{},
			want:     false,
		},
		{
			name: "chan type expression search true",
			baseExpr: &ChanTypeExpr{
				Value: &NameExpr{},
			},
			as:   &ChanTypeExpr{},
			want: true,
		},
		{
			name: "chan type expression search false",
			baseExpr: &ChanTypeExpr{
				Value: &NameExpr{},
			},
			as:   &BinaryExpr{},
			want: false,
		},
		{
			name:     "func type expression search true",
			baseExpr: &FuncTypeExpr{},
			as:       &FuncTypeExpr{},
			want:     true,
		},
		{
			name:     "func type expression search false",
			baseExpr: &FuncTypeExpr{},
			as:       &BinaryExpr{},
			want:     false,
		},
		{
			name: "map type expression search true",
			baseExpr: &MapTypeExpr{
				Key:   &NameExpr{},
				Value: &ConstExpr{},
			},
			as:   &MapTypeExpr{},
			want: true,
		},
		{
			name: "map type expression search false",
			baseExpr: &MapTypeExpr{
				Key:   &NameExpr{},
				Value: &ConstExpr{},
			},
			as:   &BinaryExpr{},
			want: false,
		},
		{
			name:     "struct type expression search true",
			baseExpr: &StructTypeExpr{},
			as:       &StructTypeExpr{},
			want:     true,
		},
		{
			name:     "struct type expression search false",
			baseExpr: &StructTypeExpr{},
			as:       &BinaryExpr{},
			want:     false,
		},
		{
			name: "const type expression search true",
			baseExpr: &constTypeExpr{
				Source: &CompositeLitExpr{},
			},
			as:   &constTypeExpr{},
			want: true,
		},
		{
			name: "const type expression search false",
			baseExpr: &constTypeExpr{
				Source: &CompositeLitExpr{},
			},
			as:   &BinaryExpr{},
			want: false,
		},
		{
			name: "maybe native type expression search true",
			baseExpr: &MaybeNativeTypeExpr{
				Type: &NameExpr{},
			},
			as:   &MaybeNativeTypeExpr{},
			want: true,
		},
		{
			name: "maybe native type expression search false",
			baseExpr: &MaybeNativeTypeExpr{
				Type: &NameExpr{},
			},
			as:   &BinaryExpr{},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ExprAsExpr(tt.baseExpr, tt.as, &seen[Expr]{})
			if got != tt.want {
				t.Errorf("ExprAsExpr() = %v, want %v", got, tt.want)
			}
		})
	}
}
