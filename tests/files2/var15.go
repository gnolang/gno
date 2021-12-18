package main

func main() {
	var xxx int
	if false {
		return
	} else if true {
		var nnn *int
		if nnn != nil {
			nnn = nil
		}
		return
	}
	println(xxx)
}
