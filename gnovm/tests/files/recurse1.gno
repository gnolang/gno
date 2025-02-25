package main

func main() {
	var x interface{}
	for i := 0; i < 10000; i++ {
		x = [1]interface{}{x}
	}
	println(x)
}

// Output:
// array[(array[(array[(array[(array[(array[(array[(array[(array[(array[(... [1]interface{})] [1]interface{})] [1]interface{})] [1]interface{})] [1]interface{})] [1]interface{})] [1]interface{})] [1]interface{})] [1]interface{})] [1]interface{})]
