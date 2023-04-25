package std

// TODO: a better encoder
func EncodeMsgCall(ChainID, PkgPath, Fn string, args []string) string {
	var as string
	for _, a := range args {
		as += a
	}
	return ChainID + "#" + PkgPath + "#" + Fn + "#" + as
}
