package main

type A struct {
	AField int
}

type B struct {
	BField int
}

type C struct {
	CField int
}

func values() (A, B, C) {
	return A{1}, B{2}, C{3}
}

func checkA(a A) {
	if a.AField != 1 {
		panic("incorrect A type")
	}
}

func checkB(b B) {
	if b.BField != 2 {
		panic("incorrect B type")
	}
}

func checkC(c C) {
	if c.CField != 3 {
		panic("incorrect C type")
	}
}

func checkOK(ok bool) {
	if !ok {
		panic("incorrect ok")
	}
}

func main() {

	// var values

	{
		var a, b, c = A{1}, B{2}, C{3}
		checkA(a)
		checkB(b)
		checkC(c)
	}

	{
		var _, b, c = A{1}, B{2}, C{3}
		checkB(b)
		checkC(c)
	}

	{
		var a, _, c = A{1}, B{2}, C{3}
		checkA(a)
		checkC(c)
	}

	{
		var _, _, c = A{1}, B{2}, C{3}
		checkC(c)
	}

	{
		var a, b, _ = A{1}, B{2}, C{3}
		checkA(a)
		checkB(b)
	}

	{
		var _, b, _ = A{1}, B{2}, C{3}
		checkB(b)
	}

	{
		var a, _, _ = A{1}, B{2}, C{3}
		checkA(a)
	}

	{
		var _, _, _ = A{1}, B{2}, C{3}
	}

	// var func

	{
		var a, b, c = values()
		checkA(a)
		checkB(b)
		checkC(c)
	}

	{
		var _, b, c = values()
		checkB(b)
		checkC(c)
	}

	{
		var a, _, c = values()
		checkA(a)
		checkC(c)
	}

	{
		var _, _, c = values()
		checkC(c)
	}

	{
		var a, b, _ = values()
		checkA(a)
		checkB(b)
	}

	{
		var _, b, _ = values()
		checkB(b)
	}

	{
		var a, _, _ = values()
		checkA(a)
	}

	{
		var _, _, _ = values()
	}

	// var type check

	{
		var ai any = A{1}
		var a, ok = ai.(A)
		checkA(a)
		checkOK(ok)
	}

	{
		var ai any = A{1}
		var _, ok = ai.(A)
		checkOK(ok)
	}

	{
		var ai any = A{1}
		var a, _ = ai.(A)
		checkA(a)
	}

	{
		var ai any = A{1}
		var _, _ = ai.(A)
	}

	// var map

	{
		var az = map[int]A{1: A{1}}
		var a, ok = az[1]
		checkA(a)
		checkOK(ok)
	}

	{
		var az = map[int]A{1: A{1}}
		var _, ok = az[1]
		checkOK(ok)
	}

	{
		var az = map[int]A{1: A{1}}
		var a, _ = az[1]
		checkA(a)
	}

	{
		var az = map[int]A{1: A{1}}
		var _, _ = az[1]
	}

	// define values

	{
		a, b, c := A{1}, B{2}, C{3}
		checkA(a)
		checkB(b)
		checkC(c)
	}

	{
		_, b, c := A{1}, B{2}, C{3}
		checkB(b)
		checkC(c)
	}

	{
		a, _, c := A{1}, B{2}, C{3}
		checkA(a)
		checkC(c)
	}

	{
		_, _, c := A{1}, B{2}, C{3}
		checkC(c)
	}

	{
		a, b, _ := A{1}, B{2}, C{3}
		checkA(a)
		checkB(b)
	}

	{
		_, b, _ := A{1}, B{2}, C{3}
		checkB(b)
	}

	{
		a, _, _ := A{1}, B{2}, C{3}
		checkA(a)
	}

	/*
		{
			_, _, _ := A{1}, B{2}, C{3}
		}
	*/

	// define func

	{
		a, b, c := values()
		checkA(a)
		checkB(b)
		checkC(c)
	}

	{
		_, b, c := values()
		checkB(b)
		checkC(c)
	}

	{
		a, _, c := values()
		checkA(a)
		checkC(c)
	}

	{
		_, _, c := values()
		checkC(c)
	}

	{
		a, b, _ := values()
		checkA(a)
		checkB(b)
	}

	{
		_, b, _ := values()
		checkB(b)
	}

	{
		a, _, _ := values()
		checkA(a)
	}

	/*
		{
			_, _, _ := values()
		}
	*/

	// define type check

	{
		var ai any = A{1}
		a, ok := ai.(A)
		checkA(a)
		checkOK(ok)
	}

	{
		var ai any = A{1}
		_, ok := ai.(A)
		checkOK(ok)
	}

	{
		var ai any = A{1}
		a, _ := ai.(A)
		checkA(a)
	}

	/*
		{
			var ai any = A{1}
			_, _ := ai.(A)
		}
	*/

	// define map

	{
		var az = map[int]A{1: A{1}}
		a, ok := az[1]
		checkA(a)
		checkOK(ok)
	}

	{
		var az = map[int]A{1: A{1}}
		_, ok := az[1]
		checkOK(ok)
	}

	{
		var az = map[int]A{1: A{1}}
		a, _ := az[1]
		checkA(a)
	}

	/*
		{
			var az = map[int]A{1: A{1}}
			_, _ := az[1]
		}
	*/
}
