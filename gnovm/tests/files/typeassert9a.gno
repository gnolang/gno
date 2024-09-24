package main

// First interface
type Reader interface {
	Read(int) string
}

type csvReader struct{}
func (r*csvReader) Read(string) string{
	return ""
}


func main() {
	var csvReader Reader = &csvReader{}
}

// Error:
// main/files/typeassert9a.gno:15:6: *main.csvReader does not implement main.Reader (wrong type for method Read)
