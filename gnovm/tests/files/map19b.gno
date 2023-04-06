package main

type cmap struct {
	servers map[int64]*server
}

type server struct {
	cm *cmap
}

func main() {
	m := cmap{}
	println(m)
}

// Output:
// struct{(nil map[int64]*main.server)}
