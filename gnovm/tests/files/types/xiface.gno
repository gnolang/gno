package main

func main() {

	s := struct {
		iface interface{}
	}{
		iface: int64(0),
	}

	switch val := s.iface.(type) {
	case int64, int:
		println(val)
		if val == 0 {
			println("l is zero")
		} else {
			println("NOT zero")
		}
	}
}

// Output:
// 0
// l is zero
