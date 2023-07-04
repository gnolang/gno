package gnolang

import (
	"fmt"
	"reflect"
	"testing"
)

type BenchValue interface {
	Int32() int32
}

type (
	Int32  int32
	Int32a int32
	Int32b int32
	Int32c int32
	Int32d int32
	Int32e int32
	Int32f int32
	Int32g int32
	Int32h int32
	Int32i int32
	Int32j int32
	Int32k int32
)

func (i Int32) Int32() int32  { return int32(i) }
func (i Int32a) Int32() int32 { return int32(i) }
func (i Int32b) Int32() int32 { return int32(i) }
func (i Int32c) Int32() int32 { return int32(i) }
func (i Int32d) Int32() int32 { return int32(i) }
func (i Int32e) Int32() int32 { return int32(i) }
func (i Int32f) Int32() int32 { return int32(i) }
func (i Int32g) Int32() int32 { return int32(i) }
func (i Int32h) Int32() int32 { return int32(i) }
func (i Int32i) Int32() int32 { return int32(i) }
func (i Int32j) Int32() int32 { return int32(i) }
func (i Int32k) Int32() int32 { return int32(i) }

func BenchmarkMapSet(b *testing.B) {
	m := make(map[int32]int32)
	for i := 0; i < b.N; i++ {
		m[int32(i%20)] = int32(i)
	}
}

func BenchmarkMapCreateSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := make(map[int32]int32)
		m[int32(i%20)] = int32(i)
	}
}

func BenchmarkMapCreateSetString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := make(map[string]int32)
		m["5"] += 1
	}
}

// shows that it might be kinda worth it to not use maps but slices for struct
// fields, but for small structs.
func BenchmarkSliceIterate10(b *testing.B) {
	fs := []TestField{}
	for i := 0; i < 10; i++ {
		fs = append(fs, TestField{fmt.Sprintf("%v", i%10), int32(i)})
	}
	for i := 0; i < b.N; i++ {
		i10 := i % 10
		for j := 0; j < 10; j++ {
			if fs[i10].Name == "5" {
				fs[i10].Value += 1
			}
		}
	}
	b.Log(fs)
}

type SomeStruct struct {
	Field1 int32
	Field2 int32
}

func (s SomeStruct) it() int32 {
	return s.Field1 + s.Field2
}

// seems to inline.
func BenchmarkStructStack(b *testing.B) {
	x := int32(0)
	for i := 0; i < b.N; i++ {
		s := SomeStruct{Field1: int32(i) % 20, Field2: 1}
		x = s.it()
	}
	b.Log(x)
}

// this doesn't work
func BenchmarkStructGC(b *testing.B) {
	x := int32(0)
	gen := func(i int) *SomeStruct { return &SomeStruct{Field1: int32(i) % 20} }
	for i := 0; i < b.N; i++ {
		s := gen(i)
		x = s.Field1
	}
	b.Log(x)
}

type TestField struct {
	Name  string
	Value int32
}

func BenchmarkTypeAssertionMethodCall(b *testing.B) {
	// This uses no interface.
	b.Run("Int32().Int32() (no interface)", func(b *testing.B) {
		var v Int32 = Int32(1)
		x := int32(0)
		for i := 0; i < b.N; i++ {
			x += v.Int32()
		}
	})
	// This calls a method on the interface.
	// It's surprising that this is slower than switch concrete assert method
	// by an order of magnitude when the alternative enables inlining.
	// Perhaps go could do better by first grouping each interface into a
	// single giant switch statement.
	b.Run("BenchValue().Int32() (interface method)", func(b *testing.B) {
		var v BenchValue = Int32(1)
		x := int32(0)
		for i := 0; i < b.N; i++ {
			x += v.Int32()
		}
	})
	// This type-asserts to a concrete type and calls its method.
	b.Run("v.(Int32).Int32() (concrete assert method)", func(b *testing.B) {
		var v interface{} = Int32(1)
		x := int32(0)
		for i := 0; i < b.N; i++ {
			x += v.(Int32).Int32()
		}
	})
	// This switch-type-asserts to a concrete type and calls its method.
	// This actually ends up being the best choice, and is even faster than
	// calling a method on an interface.
	b.Run("case v.(Int32).Int32() (type switch concrete assert method)", func(b *testing.B) {
		var v interface{} = Int32(1)
		x := int32(0)
		for i := 0; i < b.N; i++ {
			switch v := v.(type) {
			case Int32:
				x += v.Int32()
			case Int32a:
				x += v.Int32() + 1
			case Int32b:
				x += v.Int32() + 2
			case Int32c:
				x += v.Int32() + 3
			case Int32d:
				x += v.Int32() + 4
			case Int32e:
				x += v.Int32() + 5
			case Int32f:
				x += v.Int32() + 6
			case Int32g:
				x += v.Int32() + 7
			case Int32h:
				x += v.Int32() + 8
			case Int32i:
				x += v.Int32() + 9
			case Int32j:
				x += v.Int32() + 10
			case Int32k:
				x += v.Int32() + 11
			default:
				panic("should not happen")
			}
		}
	})
	// This appears to run fast, not sure what optimization is happening,
	// but maybe the initial interface setting is fine as the itable
	// info is known statically.
	b.Run("MyStruct{Value:Int32(i)} (struct interface field init)", func(b *testing.B) {
		type MyStruct struct {
			Value BenchValue
		}
		x := int32(0)
		for i := 0; i < b.N; i++ {
			s := MyStruct{Value: Int32(i)}
			x += s.Value.(Int32).Int32()
		}
	})
	// This type-asserts to an interface type and calls its method.
	// v.(BenchValue) is super slow, see https://billglover.me/2018/09/17/how-expensive-is-a-go-function-call/
	// or use `go tool compile -S test.go` for more info.
	b.Run("v.(BenchValue).Int32() (interface assert method)", func(b *testing.B) {
		var v interface{} = Int32(1)
		x := int32(0)
		for i := 0; i < b.N; i++ {
			x += v.(BenchValue).Int32()
		}
	})
}

// there is a choice between type-switching on a slice of interfaces, or to
// iterate over a slice of super-structs.
func BenchmarkTypeSwitchOrCreate(b *testing.B) {
	type Object interface{}
	type StructA struct {
		Inner Object
		A     int
		B     int
	}
	type StructB struct {
		C int
		D int
	}
	x := make([]Object, 1000)
	y := make([]StructA, 1000)
	for i := 0; i < 1000; i++ {
		x[i] = StructA{StructB{0, 0}, 0, 0}
		y[i] = StructA{StructB{0, 0}, 0, 0}
	}
	c := 0
	b.Run("type-switch", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			for i := 0; i < 1000; i++ {
				switch xi := x[i].(type) {
				case StructA:
					switch xi.Inner.(type) {
					case StructA:
						panic("shouldn't happen")
					case StructB:
						c++
					}
				case StructB:
					panic("shouldn't happen")
				}
			}
		}
		b.Log(c)
	})
	b.Run("super-struct", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			for i := 0; i < 1000; i++ {
				switch y[i].Inner.(type) {
				case StructA:
					panic("shouldn't happen")
				case StructB:
					c++
				}
			}
		}
		b.Log(c)
	})
}

func BenchmarkReflectValueOf(b *testing.B) {
	things := []interface{}{
		int(0),
		string(""),
		struct{}{},
	}
	var rv reflect.Value
	for _, thing := range things {
		b.Run(reflect.TypeOf(thing).String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				rv = reflect.ValueOf(thing)
			}
		})
	}
	b.Log(rv)
}

func BenchmarkReflectAddInt64(b *testing.B) {
	var rv reflect.Value = reflect.ValueOf(int64(1))
	var x int64
	for i := 0; i < b.N; i++ {
		x += rv.Int()
	}
	b.Log(x)
}

func BenchmarkNativeAddInt64(b *testing.B) {
	var x int64
	for i := 0; i < b.N; i++ {
		x += 1
	}
	b.Log(x)
}

func BenchmarkReflectTypeOf(b *testing.B) {
	var x int64
	var rt reflect.Type
	for i := 0; i < b.N; i++ {
		rt = reflect.TypeOf(x)
	}
	b.Log(x, rt)
}

func BenchmarkInterfaceEquality(b *testing.B) {
	ctr := 0
	var x interface{}
	var y interface{}
	for i := 0; i < b.N; i++ {
		if x == y {
			ctr++
		}
	}
	b.Log(ctr)
}

func BenchmarkPointerEquality(b *testing.B) {
	ctr := 0
	a := 1
	c := 2
	x := &a
	y := &c
	for i := 0; i < b.N; i++ {
		if x == y {
			ctr++
		}
	}
	b.Log(ctr)
}

func BenchmarkPointerDerefEquality(b *testing.B) {
	ctr := 0
	a := 1
	c := 2
	x := &a
	y := &c
	for i := 0; i < b.N; i++ {
		if *x == *y {
			ctr++
		}
	}
	b.Log(ctr)
}

func BenchmarkArrayEquality(b *testing.B) {
	b.Run("ArrayEquality[1]", func(b *testing.B) {
		ctr := 0
		x := [1]byte{0x00}
		y := [1]byte{0x00}
		for i := 0; i < b.N; i++ {
			if x == y {
				ctr++
			}
		}
		b.Log(ctr)
	})
	b.Run("ArrayEquality[8]", func(b *testing.B) {
		ctr := 0
		x := [8]byte{}
		y := [8]byte{}
		for i := 0; i < b.N; i++ {
			if x == y {
				ctr++
			}
		}
		b.Log(ctr)
	})
	b.Run("ArrayEquality[16]", func(b *testing.B) {
		ctr := 0
		x := [16]byte{}
		y := [16]byte{}
		for i := 0; i < b.N; i++ {
			if x == y {
				ctr++
			}
		}
		b.Log(ctr)
	})
	b.Run("ArrayEquality[20]", func(b *testing.B) {
		ctr := 0
		x := [20]byte{}
		y := [20]byte{}
		for i := 0; i < b.N; i++ {
			if x == y {
				ctr++
			}
		}
		b.Log(ctr)
	})
	b.Run("ArrayEquality[32]", func(b *testing.B) {
		ctr := 0
		x := [32]byte{}
		y := [32]byte{}
		for i := 0; i < b.N; i++ {
			if x == y {
				ctr++
			}
		}
		b.Log(ctr)
	})
	b.Run("ArrayEquality[64]", func(b *testing.B) {
		ctr := 0
		x := [64]byte{}
		y := [64]byte{}
		for i := 0; i < b.N; i++ {
			if x == y {
				ctr++
			}
		}
		b.Log(ctr)
	})
	b.Run("ArrayEquality[256]", func(b *testing.B) {
		ctr := 0
		x := [256]byte{}
		y := [256]byte{}
		for i := 0; i < b.N; i++ {
			if x == y {
				ctr++
			}
		}
		b.Log(ctr)
	})
}
