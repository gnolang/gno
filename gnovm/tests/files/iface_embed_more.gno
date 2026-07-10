// Interface identity is the fully-flattened method set (matching Go) across
// multi-level embedding, order, mixed embed+direct, distinctness, the empty
// interface, and dynamic dispatch through an embed.
package main

type Str interface{ S() string }
type Rdr interface{ R() int }
type Wtr interface{ W() int }
type L1 interface{ Str }
type L2 interface{ L1 } // multi-level

type T struct{}

func (T) S() string { return "s" }

func eq(a, b interface{}) bool { return a == b }

func main() {
	println("multi-level:", eq(struct{ X interface{ L2 } }{}, struct{ X interface{ S() string } }{}))
	println("order:", eq(struct {
		X interface {
			Rdr
			Wtr
		}
	}{}, struct {
		X interface {
			Wtr
			Rdr
		}
	}{}))
	println("mixed:", eq(struct {
		X interface {
			Str
			R() int
		}
	}{}, struct {
		X interface {
			S() string
			Rdr
		}
	}{}))
	println("superset!=subset:", eq(struct {
		X interface {
			Rdr
			Wtr
		}
	}{}, struct{ X interface{ Rdr } }{}))
	println("empty-embed:", eq(struct{ X interface{ interface{} } }{}, struct{ X interface{} }{}))
	var v interface{} = T{}
	s, ok := v.(interface{ Str })
	if ok {
		println("dispatch:", s.S())
	} else {
		println("dispatch: !ok")
	}
}

// Output:
// multi-level: true
// order: true
// mixed: true
// superset!=subset: false
// empty-embed: true
// dispatch: s
