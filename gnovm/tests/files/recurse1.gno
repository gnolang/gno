package main

func main() {
	var x any
	for i := 0; i < 10000; i++ {
		x = [1]any{x}
	}
	println(x)
}

// Output:
// array[(array[(array[(array[(array[(array[(array[(array[(array[(array[(... [1]interface {})] [1]interface {})] [1]interface {})] [1]interface {})] [1]interface {})] [1]interface {})] [1]interface {})] [1]interface {})] [1]interface {})] [1]interface {})]
