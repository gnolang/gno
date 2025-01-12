package main

import (
	"fmt"
	"time"
)

func main() {
	const a int = 1_000_000
	const b byte = byte(1)
	const c float64 = 1_000_000.000
	const d string = "Hello, World!"
	const e rune = 'a'
	const g bool = true
	const h uint = 1_000
	const i int8 = 1
	const j int16 = 1
	const k int32 = 1
	const l int64 = 1
	const m uint8 = 1
	const n uint16 = 1
	const o uint32 = 1
	const p uint64 = 1
	const r float32 = 1_000_000.000
	const s = r
	const t = len("s")
	const u = 1 + len("s") + 3
	ars := [10]string{}
	const v = len(ars)
	const w = cap(ars)
	const x = time.Second
	const y = +len("ay")

	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)
	fmt.Println(d)
	fmt.Println(e)
	fmt.Println(g)
	fmt.Println(h)
	fmt.Println(i)
	fmt.Println(j)
	fmt.Println(k)
	fmt.Println(l)
	fmt.Println(m)
	fmt.Println(n)
	fmt.Println(o)
	fmt.Println(p)
	fmt.Println(r)
	fmt.Println(s)
	fmt.Println(t)
	fmt.Println(u)
	fmt.Println(v)
	fmt.Println(w)
	println(x)
	fmt.Println(y)
}

// Output:
// 1000000
// 1
// 1e+06
// Hello, World!
// 97
// true
// 1000
// 1
// 1
// 1
// 1
// 1
// 1
// 1
// 1
// 1e+06
// 1e+06
// 1
// 5
// 10
// 10
// 1s
// 2
