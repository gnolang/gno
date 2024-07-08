package tests

import (
	"fmt"
	"reflect"
	"testing"
)

/*
This attempts to show a sufficiently exhaustive list of ValuePaths for
different types of selectors.  As can be seen, even a simple selector
expression can represent a number of different types of selectors.
*/

// S1 struct
type S1 struct {
	F0 int
}

func (S1) Hello() {
}

func (*S1) Bye() {
}

// Pointer to S1
type S1P *S1

// Like S1 but pointer struct
type PS1 *struct {
	F0 int
}

type S7 struct {
	S1
}

type S9 struct {
	*S1
}

type S10PD *struct {
	S1
}

func _printValue(x interface{}) {
	if reflect.TypeOf(x).Kind() == reflect.Func {
		fmt.Println("function")
	} else {
		fmt.Println(x)
	}
}

func TestSelectors(t *testing.T) {
	t.Parallel()

	x0 := struct{ F0 int }{1}
	_printValue(x0.F0) //       *ST.F0
	//                            F:0
	//                          VPField{depth:0,index:0}
	x1 := S1{1}
	_printValue(x1.F0) //       *DT(S1)>*ST.F0
	//                          +1        F:0
	//                          VPField{depth:1,index:0}
	_printValue(x1.Hello) //    *DT(S1).Hello
	//                          +1    M:0
	//                          VPValMethod{index:0}
	_printValue(x1.Bye) //      *PT(implied)>*DT(S1).Bye
	//                          +D           +1   *M:1
	//                          VPDerefPtrMethod{index:1}
	x2 := &x0
	_printValue(x2.F0) //       *PT>*ST.F0
	//                          +D    F:0
	//                          VPDerefField{depth:0,index:0}
	var x3 PS1 = &struct{ F0 int }{1}
	_printValue(x3.F0) //       *DT(S1P)>*PT>*ST.F0
	//                          +1       +D   F:0
	//                          VPDerefField{depth:1,index:0}
	x4 := &S1{1}
	_printValue(x4.F0) //       *PT>*DT(S1P)>*ST.F0
	//                          +D  +1         F:0
	//                          VPDerefField{depth:2,index:0}
	var x5 S1P = &S1{1}
	_printValue(x5.F0) //       *DT(S1P)>*PT>*DT(S1)>*ST.F0
	//                          +1       +D  +1        F:0
	//                          VPDerefField{depth:3,index:0}
	x6 := &x5
	_printValue(x6)
	// _printValue(x6.F0)       *PT>*DT(S1P)??? > *PT>*DT(S1)>*ST.F0
	//                          +D  +1            +D  +1        F:0
	//                          VPDerefField{depth:1,index:0}(WRONG!!!) > VPDerefField{depth:1,index:0} XXX ERROR
	x7 := S7{S1{1}}
	_printValue(x7.F0) //       *DT(S7)>*ST.S1 > *DT(S1)>*ST.F0
	//                          +1        F:0    +1        F:0
	//                          VPField{depth:1,index:0} > VPField{depth:1,index:0}
	x8 := &x7
	_printValue(x8.F0) //       *PT>*DT(S7)>*ST.S1 > *DT(S1)>*ST.F0
	//                          +D  +1        F:0    +1        F:0
	//                          VPDerefField{depth:1,index:0} > VPField{depth:1,index:0}
	x9 := S9{x5}
	_printValue(x9.F0) //       *DT(S9)>*ST.S1 > *PT>*DT(S1)>*ST.F0
	//                          +1        F:0    +D  +1        F:0
	//                          VPField{depth:1,index:0} > VPDerefField{depth:1,index:0}
	x10 := struct{ S1 }{S1{1}}
	_printValue(x10.F0) //      *ST.S1 > *DT(S1)>*ST.F0
	//                            F:0    +1        F:0
	//                          VPField{depth:0,index:0} > VPField{depth:1,index:0}
	_printValue(x10.Hello) //   *ST.S1 > *DT(S1).Hello
	//                            F:0    +1    M:0
	//                          VPField{depth:0,index:0} > VPValMethod{index:0}
	_printValue(x10.Bye) //     (*PT>)*ST.S1 > *DT(S1).Bye
	//                           +S     F:0    +1   *M:1
	//                          VPSubrefField{depth:0,index:0} > VPDerefPtrMethod{index:1}
	x10p := &x10
	_printValue(x10p.F0) //     *PT>*ST.S1 > *DT(S1)>*ST.F0
	//                          +D    F:0    +1        F:0
	//                          VPDerefField{depth:0,index:0} > VPField{depth:1,index:0}
	_printValue(x10p.Hello) //  *PT>*ST.S1 > *DT(S1).Hello
	//                          +D    F:0    +1    M:0
	//                          VPDerefField{depth:0,index:0} > VPValMethod{index:0}
	_printValue(x10p.Bye) //    *PT>*ST.S1 > *DT(S1).Bye
	//                          +D    F:0    +1   *M:1
	//                          VPSubrefField{depth:0,index:0} > VPDerefPtrMethod{index:1}
	var x10pd S10PD = &struct{ S1 }{S1{1}}
	_printValue(x10pd.F0) //    *DT(S10PD)>*PT>*ST.S1 > *DT(S1)>*ST.F0
	//                          +1         +D    F:0    +1        F:0
	//                          VPDerefField{depth:1,index:0} > VPField{depth:1,index:0}
	// _printValue(x10pd.Hello) *DT(S10PD)>*PT>*ST.S1 > *DT(S1).Hello XXX weird, doesn't work.
	//                          +1         +D    F:0    +1    M:0
	//                          VPDerefField{depth:1,index:0} > VPValMethod{index:0}
	_printValue(x10p.Bye) //    *DT(S10PD)>*PT>*ST.S1 > *DT(S1).Bye
	//                          +1         +D    F:0    +1   *M:1
	//                          VPSubrefField{depth:1,index:0} > VPDerefPtrMethod{index:1}
	x11 := S7{S1{1}}
	_printValue(x11.F0) //      *DT(S7)>*ST.S1 > *DT(S1)>*ST.F0 NOTE same as x7.
	//                          +1        F:0    +1        F:0
	//                          VPField{depth:1,index:0} > VPField{depth:1,index:0}
	_printValue(x11.Hello) //   *DT(S7)>*ST.S1 > *DT(S1)>*ST.Hello
	//                          +1        F:0    +1        M:0
	//                          VPField{depth:1,index:0} > VPValMethod{index:0}
	_printValue(x11.Bye) //     (*PT>)*DT(S7)>*ST.S1 > *DT(S1).Bye
	//                           +S   +1        F:0    +1   *M:1
	//                          VPSubrefField{depth:2,index:0} > VPDerefPtrMethod{index:1}
	x11p := &S7{S1{1}}
	_printValue(x11p.F0) //     *PT>*DT(S7)>*ST.S1 > *DT(S1)>*ST.F0
	//                          +1            F:0    +1        F:0
	//                          VPDerefField{depth:2,index:0} > VPField{depth:1,index:0}
	_printValue(x11p.Hello) //  *PT>*DT(S7)>*ST.S1 > *DT(S1).Hello
	//                          +1            F:0    +1    M:0
	//                          VPDerefField{depth:2,index:0} > VPValMethod{index:0}
	_printValue(x11p.Bye) //    *PT>*DT(S7)>*ST.S1 > *DT(S1).Bye
	//                          +1            F:0    +1   *M:1
	//                          VPSubrefField{depth:2,index:0} > VPDerefPtrMethod{index:1}
	x12 := struct{ *S1 }{&S1{1}}
	_printValue(x12.F0) //      *ST.S1 > *PT>*DT(S1)>*ST.F0
	//                            F:0    +D  +1        F:0
	//                          VPField{depth:0,index:0} > VPDerefField{depth:1,index:0}
	_printValue(x12.Hello) //   *ST.S1 > *PT>*DT(S1).Hello
	//                            F:0    +D  +1    M:0
	//                          VPField{depth:0,index:0} > VPDerefValMethod{index:0}
	_printValue(x12.Bye) //     *ST.S1 > *PT>*DT(S1).Bye
	//                            F:0    +D  +1   *M:1
	//                          VPField{depth:0,index:0} > VPDerefPtrMethod{index:1}
	x13 := &x12
	_printValue(x13.F0) //      *PT>*ST.S1 > *PT>*DT(S1)>*ST.F0
	//                          +D    F:0    +D  +1        F:0
	//                          VPDerefField{depth:0,index:0} > VPDerefField{depth:1,index:0}
	_printValue(x13.Hello) //   *PT>*ST.S1 > *PT>*DT(S1).Hello
	//                          +D    F:0    +D  +1    M:0
	//                          VPDerefField{depth:0,index:0} > VPDerefValMethod{index:0}
	_printValue(x13.Bye) //     *PT>*ST.S1 > *PT>*DT(S1).Bye
	//                          +D    F:0    +D  +1   *M:1
	//                          VPDerefField{depth:0,index:0} > VPDerefPtrMethod{index:1}
}
