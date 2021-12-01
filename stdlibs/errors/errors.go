package errors

type stringError string

func (serr stringError) Error() string {
	return string(serr)
}

// TODO: implement runtime.Caller and fmt.Sprintf,
// and port pkgs/errors/errors.go as default "errors" impl.
func New(msg string) stringError {
	return stringError(msg)
}
