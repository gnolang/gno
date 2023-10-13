package std

// TODO: a better encoder
func EncodeMsgCall(ChainID, PkgPath, Fn string, args []string) string {
	var as string
	for i, a := range args {
		as += a
		if i != (len(args) - 1) {
			as += "&"
		}
	}
	return ChainID + "#" + PkgPath + "#" + Fn + "#" + as
}
