package main

import (
	"fmt"
	"reflect"
	"testing"
)

type Bar struct {
	B int
	C int
	A string
}

type Foo struct {
	B   int    // 0
	C   string // 1
	D   int64  // 2
	E   int    // 3
	X   *Bar   // 4
	Y   Bar    // 5
	Bar        // 6
}

//----------------------------------------
// 1

func BenchmarkReflectFieldIndex(b *testing.B) {
	f := Foo{X: &Bar{A: "thisissomerandomstringthatiskindalongenough"}, Y: Bar{A: "thisissomerandomstringthatiskindalongenough"}, Bar: Bar{A: "thisissomerandomstringthatiskindalongenough"}}
	rv := reflect.ValueOf(f)
	t := int(0)
	for i := 0; i < b.N; i++ {
		frv := rv.Field(5)
		frv = frv.Field(2)
		t += frv.Len()
	}
	fmt.Println(t)
}

func BenchmarkReflectFieldName(b *testing.B) {
	f := Foo{X: &Bar{A: "thisissomerandomstringthatiskindalongenough"}, Y: Bar{A: "thisissomerandomstringthatiskindalongenough"}, Bar: Bar{A: "thisissomerandomstringthatiskindalongenough"}}
	rv := reflect.ValueOf(f)
	t := int(0)
	for i := 0; i < b.N; i++ {
		frv := rv.FieldByName("Y")
		frv = frv.FieldByName("A")
		t += frv.Len()
	}
	fmt.Println(t)
}

func GetA(f Foo) string {
	return string(f.Y.A)
}

func BenchmarkFunctionField(b *testing.B) {
	f := Foo{X: &Bar{A: "thisissomerandomstringthatiskindalongenough"}, Y: Bar{A: "thisissomerandomstringthatiskindalongenough"}, Bar: Bar{A: "thisissomerandomstringthatiskindalongenough"}}
	t := int(0)
	for i := 0; i < b.N; i++ {
		t += len(GetA(f))
	}
	fmt.Println(t)
}

//----------------------------------------
// 2

func BenchmarkReflectFieldIndex2(b *testing.B) {
	f := Foo{X: &Bar{A: "thisissomerandomstringthatiskindalongenough"}, Y: Bar{A: "thisissomerandomstringthatiskindalongenough"}, Bar: Bar{A: "thisissomerandomstringthatiskindalongenough"}}
	rv := reflect.ValueOf(f)
	t := int(0)
	for i := 0; i < b.N; i++ {
		frv := rv.Field(6)
		frv = frv.Field(2)
		t += frv.Len()
	}
	fmt.Println(t)
}

func BenchmarkReflectFieldName2(b *testing.B) {
	f := Foo{X: &Bar{A: "thisissomerandomstringthatiskindalongenough"}, Y: Bar{A: "thisissomerandomstringthatiskindalongenough"}, Bar: Bar{A: "thisissomerandomstringthatiskindalongenough"}}
	rv := reflect.ValueOf(f)
	t := int(0)
	for i := 0; i < b.N; i++ {
		frv := rv.FieldByName("A")
		t += frv.Len()
	}
	fmt.Println(t)
}

func GetA2(f Foo) string {
	return string(f.A)
}

func BenchmarkFunctionField2(b *testing.B) {
	f := Foo{X: &Bar{A: "thisissomerandomstringthatiskindalongenough"}, Y: Bar{A: "thisissomerandomstringthatiskindalongenough"}, Bar: Bar{A: "thisissomerandomstringthatiskindalongenough"}}
	t := int(0)
	for i := 0; i < b.N; i++ {
		t += len(GetA2(f))
	}
	fmt.Println(t)
}

//----------------------------------------
// 3

func BenchmarkReflectFieldIndex3(b *testing.B) {
	f := Foo{X: &Bar{A: "thisissomerandomstringthatiskindalongenough"}, Y: Bar{A: "thisissomerandomstringthatiskindalongenough"}, Bar: Bar{A: "thisissomerandomstringthatiskindalongenough"}}
	rv := reflect.ValueOf(f)
	t := int(0)
	for i := 0; i < b.N; i++ {
		frv := rv.Field(4)
		frv = frv.Elem().Field(2)
		t += frv.Len()
	}
	fmt.Println(t)
}

func BenchmarkReflectFieldName3(b *testing.B) {
	f := Foo{X: &Bar{A: "thisissomerandomstringthatiskindalongenough"}, Y: Bar{A: "thisissomerandomstringthatiskindalongenough"}, Bar: Bar{A: "thisissomerandomstringthatiskindalongenough"}}
	rv := reflect.ValueOf(f)
	t := int(0)
	for i := 0; i < b.N; i++ {
		frv := rv.FieldByName("X")
		frv = frv.Elem().FieldByName("A")
		t += frv.Len()
	}
	fmt.Println(t)
}

func GetA3(f Foo) string {
	return string(f.X.A)
}

func BenchmarkFunctionField3(b *testing.B) {
	f := Foo{X: &Bar{A: "thisissomerandomstringthatiskindalongenough"}, Y: Bar{A: "thisissomerandomstringthatiskindalongenough"}, Bar: Bar{A: "thisissomerandomstringthatiskindalongenough"}}
	t := int(0)
	for i := 0; i < b.N; i++ {
		t += len(GetA3(f))
	}
	fmt.Println(t)
}
