package main

func main() {
	dst := make([]struct {
		A int `a`
	}, 1)
	src := []struct {
		A int `b`
	}{{1}}
	copy(dst, src)
}

// TypeCheckError:
// main/copy_struct_tags.gno:10:7: invalid copy: arguments dst (variable of type []struct{A int "a"}) and src (variable of type []struct{A int "b"}) have different element types struct{A int "a"} and struct{A int "b"}
