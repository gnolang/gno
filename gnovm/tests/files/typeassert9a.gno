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
// main/typeassert9a.gno:15:6-37: *main.csvReader does not implement main.Reader (wrong type for method Read)

// TypeCheckError:
// main/typeassert9a.gno:15:25: cannot use &csvReader{} (value of type *csvReader) as Reader value in variable declaration: *csvReader does not implement Reader (wrong type for method Read)
// 		have Read(string) string
// 		want Read(int) string; main/typeassert9a.gno:15:6: declared and not used: csvReader
