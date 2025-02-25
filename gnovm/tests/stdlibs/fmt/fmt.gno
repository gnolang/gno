package fmt

func Println(a ...interface{}) (int, error) {
	return 0, nil
}

func Printf(format string, a ...interface{}) (int, error) {
	return 0, nil
}

func Print(a ...interface{}) (int, error) {
	return 0, nil
}

func Sprint(a ...interface{}) string {
	return ""
}

func Sprintf(format string, a ...interface{}) string {
	return ""
}

func Sprintln(a ...interface{}) string {
	return ""
}

func Sscanf(str string, format string, a ...interface{}) (n int, err error) {
	return 0, nil
}

func Errorf(format string, a ...interface{}) error {
	return &errorFmt{err: format}
}

func Fprintln(w Writer, a ...interface{}) (n int, err error) {
	return 0, nil
}

func Fprintf(w Writer, format string, a ...interface{}) (n int, err error) {
	return 0, nil
}

func Fprint(w Writer, a ...interface{}) (n int, err error) {
	return 0, nil
}

type errorFmt struct {
	err string
}

func (e *errorFmt) Error() string {
	return e.err
}

type Writer interface {
	Write(p []byte) (n int, err error)
}
