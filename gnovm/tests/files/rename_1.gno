package main

func main() {
	for i:=0; i<3; i++{
		println("start: ", i)
		{
			i := i+1
			println("shadow: ", i) // should not rename.
		}
		println("end: ", i)
	}
}