package std

// TODO: a better encoder
func EncodeMsgCall(PkgPath, Fn string, args []string) string {
	var as string
	for _, a := range args {
		as += a
	}
	return PkgPath+"#"+Fn+"#"+as
}
