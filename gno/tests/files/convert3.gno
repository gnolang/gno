package main

func fakeSplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	return 7, nil, nil
}

func SplitFunc(fn func([]byte, bool) (int, []byte, error)) func([]byte, bool) (int, []byte, error) {
	return func(data []byte, atEOF bool) (int, []byte, error) {
		return fn(data, atEOF)
	}
}

func main() {
	splitfunc := SplitFunc(fakeSplitFunc)
	n, _, err := splitfunc(nil, true)
	if err != nil {
		panic(err)
	}
	println(n)
}

// Output:
// 7
